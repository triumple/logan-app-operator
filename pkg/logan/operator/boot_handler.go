package operator

import (
	"fmt"
	"github.com/go-logr/logr"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/logancloud/logan-app-operator/pkg/logan/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
)

const (
	DefaultAppName               = "app"
	DefaultImagePullPolicy       = "Always"
	DefaultRevisionHistoryLimits = int(5)
	DefaultWeight                = 100

	EnvAnnotationKey        = "app.logancloud.com/env"
	EnvAnnotationValue      = "generated"
	BootEnvsAnnotationKey   = "app.logancloud.com/boot-envs"
	BootImagesAnnotationKey = "app.logancloud.com/boot-images"

	DeployAnnotationKey   = "app.logancloud.com/deploy"
	ServicesAnnotationKey = "app.logancloud.com/services"
	AppTypeAnnotationKey  = "app.logancloud.com/type"

	MutationAnnotationKey = "app.logancloud.com/mutation"

	EventTypeNormal  = "Normal"
	EventTypeWarning = "Warning"
)

type BootHandler struct {
	OperatorBoot metav1.Object
	OperatorSpec *appv1.BootSpec
	OperatorMeta *metav1.ObjectMeta

	Boot   *appv1.Boot
	Config *config.BootConfig

	Scheme   *runtime.Scheme
	Client   client.Client
	Logger   logr.Logger
	Recorder record.EventRecorder
}

func (handler *BootHandler) UpdateAnnotation(annotationMap map[string]string) bool {
	metaData := handler.OperatorMeta
	updated := false

	if metaData.Annotations == nil {
		metaData.Annotations = make(map[string]string)
	}

	for aKey, aValue := range annotationMap {
		if metaData.Annotations[aKey] != aValue {
			metaData.Annotations[aKey] = aValue
			updated = true
		}
	}

	return updated
}

// NewDeployment return a new created Boot's Deployment object
func (handler *BootHandler) NewDeployment() *appsv1.Deployment {
	logger := handler.Logger
	boot := handler.Boot
	bootCfg := handler.Config

	revisionHistoryLimits := int32(DefaultRevisionHistoryLimits)
	podLabels := PodLabels(boot)
	deployLabels := DeployLabels(boot)

	containers := []corev1.Container{*handler.NewAppContainer()}

	sidecarContainers := bootCfg.SidecarContainers

	if sidecarContainers != nil {
		for _, c := range *sidecarContainers {
			sideCarContainer := c.DeepCopy()
			// Replace Envs
			DecodeEnvs(boot, sideCarContainer.Env)

			containers = append(containers, *sideCarContainer)
		}
	}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeployName(boot),
			Namespace: boot.Namespace,
			Labels:    deployLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             boot.Spec.Replicas,
			RevisionHistoryLimit: &revisionHistoryLimits,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: DefaultWeight,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      boot.AppKey,
													Operator: "In",
													Values:   []string{boot.Name},
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					Containers:   containers,
					NodeSelector: boot.Spec.NodeSelector,
				},
			},
			Strategy: appsv1.DeploymentStrategy{},
		},
	}

	// Avoid when boot has more than 4 pods, more than one pod will be RollingUpdate.
	if boot.BootType == logan.BootJava {
		maxUnavailable := intstr.FromString("1%")
		dep.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: &maxUnavailable,
			},
		}
	}

	podSpec := bootCfg.AppSpec.PodSpec
	if podSpec != nil {
		appPodSpec := *podSpec.DeepCopy()
		err := util.MergeOverride(&dep.Spec.Template.Spec, appPodSpec)
		if err != nil {
			logger.Error(err, "config merge error.", "type", "podSpec")
		}

		initContainers := dep.Spec.Template.Spec.InitContainers
		if initContainers != nil && len(initContainers) > 0 {
			for _, c := range initContainers {
				DecodeEnvs(boot, c.Env)
			}
		}
	}

	_ = controllerutil.SetControllerReference(handler.OperatorBoot, dep, handler.Scheme)

	return dep
}

// NewAppContainer return a new created App Container instance
func (handler *BootHandler) NewAppContainer() *corev1.Container {
	boot := handler.Boot
	healthPort := AppContainerHealthPort(boot, handler.Config.AppSpec)
	imageName := AppContainerImageName(boot, handler.Config.AppSpec)

	appContainer := corev1.Container{
		Image: imageName,
		Name:  DefaultAppName,
		Ports: []corev1.ContainerPort{{
			ContainerPort: boot.Spec.Port,
			Name:          HttpPortName,
		}},
		Env:             boot.Spec.Env,
		ImagePullPolicy: DefaultImagePullPolicy,
		Resources:       boot.Spec.Resources,

		LivenessProbe: &corev1.Probe{
			FailureThreshold: 10,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   boot.Spec.Health,
					Port:   healthPort,
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 120,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold: 10,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   boot.Spec.Health,
					Port:   healthPort,
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 60,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      5,
		},
	}

	if boot.Spec.Command != "" {
		appContainer.Command = []string{boot.Spec.Command}
	}

	specContainer := handler.Config.AppSpec.Container
	if specContainer != nil {
		err := util.MergeOverride(&appContainer, *specContainer)
		if err != nil {
			handler.Logger.Error(err, "Merge error.", "type", "container")
		}
	}

	return &appContainer
}

// NewService returns a new created Service instance
func (handler *BootHandler) NewServices() []*corev1.Service {
	boot := handler.Boot
	bootCfg := handler.Config
	// app Service
	bootSvc := handler.createService(int(boot.Spec.Port), boot.Name)
	allSvcs := []*corev1.Service{bootSvc}

	// additional sidecar Service
	sidecarSvcs := bootCfg.SidecarServices
	if sidecarSvcs != nil {
		for _, svc := range *sidecarSvcs {
			svcName, _ := Decode(boot, svc.Name)
			allSvcs = append(allSvcs, handler.createService(int(svc.Port), svcName))
		}
	}

	return allSvcs
}

// NewService returns a new created Service instance
func (handler *BootHandler) createService(port int, name string) *corev1.Service {
	boot := handler.Boot

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceName(boot, name),
			Namespace:   boot.Namespace,
			Labels:      ServiceLabels(boot),
			Annotations: ServiceAnnotation(port),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       HttpPortName,
					Port:       int32(port),
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: int32(port)},
				},
			},
			Selector: PodLabels(boot),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	// Set Boot instance as the owner and controller
	_ = controllerutil.SetControllerReference(handler.OperatorBoot, svc, handler.Scheme)

	return svc
}

func (handler *BootHandler) EventNormal(reason string, obj string) {
	recorder := handler.Recorder
	boot := handler.Boot

	recorder.Event(boot, EventTypeNormal, reason, fmt.Sprintf("Successful: obj=%s", obj))
}

func (handler *BootHandler) EventFail(reason string, obj string, err error) {
	recorder := handler.Recorder
	boot := handler.Boot

	recorder.Event(boot, EventTypeWarning, reason, fmt.Sprintf("Failed: obj=%s, err=%s", obj, err.Error()))
}
