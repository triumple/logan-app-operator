package operator

import (
	"context"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"strings"
)

// ReconcileCreate check the existence of components, if not exist, create new one.
// 1. Deployment not found: Create Deployment, requeue=false
// 2. Service not found: Create Service, requeue=false
// 3. When creating Error: requeue error requeue=true
func (handler *BootHandler) ReconcileCreate() (reconcile.Result, error, bool) {
	boot := handler.Boot
	logger := handler.Logger
	c := handler.Client

	depFound := &appsv1.Deployment{}
	depName := DeployName(boot)
	err := c.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: boot.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		reason := "Creating Deployment"
		if errors.IsNotFound(err) {
			dep := handler.NewDeployment()
			logger.Info(reason, "deploy containers", dep.Spec.Template.Spec.Containers)
			err = c.Create(context.TODO(), dep)
			if err != nil {
				logger.Error(err, "Failed to create Deployment", "deploy", depName)
				handler.EventFail(reason, dep.Name, err)
				return reconcile.Result{}, err, true
			}
			handler.EventNormal(reason, dep.Name)
		} else {
			logger.Error(err, "Failed to get Deployment")
			handler.EventFail(reason, depName, err)
			return reconcile.Result{}, err, true
		}
	}

	appSvcFound := &corev1.Service{}
	appSvcName := ServiceName(boot, boot.Name)
	err = c.Get(context.TODO(), types.NamespacedName{Name: appSvcName, Namespace: boot.Namespace}, appSvcFound)
	if err != nil && errors.IsNotFound(err) {
		reason := "Creating Service"
		if errors.IsNotFound(err) {
			// App Service not found, create app Service and sidecar services if necessary.
			// need to consider individually in the future

			// Creating all services
			for _, svc := range handler.NewServices() {
				logger.Info("Creating Service", "service", svc.Name)
				err = c.Create(context.TODO(), svc)
				if err != nil {
					//Note: Maybe when it called, the service is not created yet.
					logger.Info("Failed to create new Service, maybe the service is not created successfully",
						"service", svc.Name)
					handler.EventFail(reason, svc.Name, err)
					return reconcile.Result{}, nil, true
				}
				handler.EventNormal(reason, svc.Name)
			}
		} else {
			logger.Error(err, "Failed to get Services")
			handler.EventFail(reason, boot.Name, err)
			return reconcile.Result{}, err, true
		}
	}

	return reconcile.Result{}, nil, false
}

// ReconcileUpdate check the fields of components, if not as desire, update it.
// 1. Check Deployment's existence: error -> requeue=true
// 1.1. Check Deployment's fields: "replicas", image, env, port, resources, health, nodeSelector
// 2. Check Service's existence: error -> requeue=true
// 2.1 Check Service's fields:
func (handler *BootHandler) ReconcileUpdate() (reconcile.Result, error, bool) {
	boot := handler.Boot
	logger := handler.Logger
	c := handler.Client

	//1 Deployment
	depFound := &appsv1.Deployment{}
	depName := DeployName(boot)
	err := c.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: boot.Namespace}, depFound)
	if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return reconcile.Result{Requeue: true}, err, true
	}
	result, err, requeue := handler.reconcileUpdateDeploy(depFound)
	if requeue {
		return result, err, true
	}

	//2 Service
	appSvcFound := &corev1.Service{}
	appSvcName := ServiceName(boot, boot.Name)
	err = c.Get(context.TODO(), types.NamespacedName{Name: appSvcName, Namespace: boot.Namespace}, appSvcFound)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Service resource not found. Ignoring since object is not created successfully yet", "error", err)
			return reconcile.Result{}, nil, true
		}
		logger.Error(err, "Failed to get Service")
		return reconcile.Result{Requeue: true}, err, true
	}
	result, err, requeue = handler.reconcileUpdateService(appSvcFound)
	if requeue {
		return result, err, true
	}

	return reconcile.Result{}, nil, false
}

// reconcileUpdateDeploy handle update logic of Deployment
func (handler *BootHandler) reconcileUpdateDeploy(deploy *appsv1.Deployment) (reconcile.Result, error, bool) {
	logger := handler.Logger
	boot := handler.Boot
	c := handler.Client

	updated := false

	reason := "Updating Deployment"
	// 1. Check ownerReferences
	ownerReferences := deploy.OwnerReferences
	if ownerReferences == nil || len(ownerReferences) == 0 {
		logger.Info(reason, "type", "ownerReferences", "deploy", deploy.Name)

		_ = controllerutil.SetControllerReference(handler.OperatorBoot, deploy, handler.Scheme)

		updated = true
	}

	// 2. Check size
	size := boot.Spec.Replicas
	if *deploy.Spec.Replicas != *size {
		logger.Info(reason, "type", "replicas", "deploy", deploy.Name,
			"old", deploy.Spec.Replicas, "new", size)
		*deploy.Spec.Replicas = *size

		updated = true
	}

	// "spec.template.spec.containers" is a required value, no need to verify.
	// 3. Check image and version:
	deployImg := deploy.Spec.Template.Spec.Containers[0].Image
	bootImg := AppContainerImageName(handler.Boot, handler.Config.AppSpec)
	if bootImg != deployImg {
		logger.Info(reason, "type", "image", "Deploy", deploy.Name,
			"old", deployImg, "new", bootImg)
		deploy.Spec.Template.Spec.Containers[0].Image = bootImg

		updated = true
	}

	// 4. Check env: check fist container(boot container)
	deployEnv := deploy.Spec.Template.Spec.Containers[0].Env
	bootEnv := boot.Spec.Env
	if !reflect.DeepEqual(deployEnv, bootEnv) {
		logger.Info(reason, "type", "env", "deploy", deploy.Name,
			"old", deployEnv, "new", bootEnv)
		deploy.Spec.Template.Spec.Containers[0].Env = bootEnv

		updated = true
	}

	// 5. Check port: check fist container(boot container)
	deployPorts := deploy.Spec.Template.Spec.Containers[0].Ports
	bootPorts := []corev1.ContainerPort{{Name: HttpPortName, ContainerPort: boot.Spec.Port, Protocol: corev1.ProtocolTCP}}
	if !reflect.DeepEqual(deployPorts, bootPorts) {
		logger.Info(reason, "type", "port", "deploy", deploy.Name,
			"old", deployPorts, "new", bootPorts)
		deploy.Spec.Template.Spec.Containers[0].Ports = bootPorts
		readinessProbe := deploy.Spec.Template.Spec.Containers[0].ReadinessProbe
		if readinessProbe != nil {
			readinessProbe.HTTPGet.Port = intstr.IntOrString{Type: intstr.Int, IntVal: int32(boot.Spec.Port)}
		}

		livenessProbe := deploy.Spec.Template.Spec.Containers[0].LivenessProbe
		if readinessProbe != nil {
			livenessProbe.HTTPGet.Port = intstr.IntOrString{Type: intstr.Int, IntVal: int32(boot.Spec.Port)}
		}

		updated = true
	}

	// 6 Check resources: check fist container(boot container)
	deployResources := deploy.Spec.Template.Spec.Containers[0].Resources
	bootResources := boot.Spec.Resources
	if !reflect.DeepEqual(deployResources, bootResources) {
		logger.Info(reason, "type", "resources", "deploy", deploy.Name,
			"old", deployResources, "new", bootResources)
		deploy.Spec.Template.Spec.Containers[0].Resources = bootResources

		updated = true
	}

	// 7 Check health: check fist container(boot container)
	probe := deploy.Spec.Template.Spec.Containers[0].LivenessProbe
	bootHealth := *boot.Spec.Health
	if bootHealth == "" {
		if probe != nil {
			// Remove the 2 existing probes.
			deployHealth := probe.HTTPGet.Path
			logger.Info(reason, "type", "health", "deploy", deploy.Name,
				"old", deployHealth, "new", "")
			deploy.Spec.Template.Spec.Containers[0].LivenessProbe = nil
			deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = nil
			updated = true
		}
	} else {
		if probe == nil {
			// 1. If probe is nil, add Liveness and Readiness
			liveness, readiness := handler.GetHealthProbe()
			logger.Info(reason, "type", "health", "deploy", deploy.Name,
				"old", "empty", "new", bootHealth)
			deploy.Spec.Template.Spec.Containers[0].LivenessProbe = liveness
			deploy.Spec.Template.Spec.Containers[0].ReadinessProbe = readiness
			updated = true
		} else {
			deployHealth := probe.HTTPGet.Path
			// 2. If probe is not nil, we only need to update the health path
			if deployHealth != bootHealth {
				logger.Info(reason, "type", "health", "deploy", deploy.Name,
					"old", deployHealth, "new", bootHealth)
				deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path = bootHealth
				deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path = bootHealth
				updated = true
			}
		}
	}

	// 8 Check nodeSelector: map[string]string
	deployNodeSelector := deploy.Spec.Template.Spec.NodeSelector
	bootNodeSelector := boot.Spec.NodeSelector
	if !reflect.DeepEqual(deployNodeSelector, bootNodeSelector) {
		logger.Info(reason, "type", "nodeSelector", "deploy", deploy.Name,
			"old", deployNodeSelector, "new", bootNodeSelector)
		deploy.Spec.Template.Spec.NodeSelector = bootNodeSelector

		updated = true
	}

	// 9 Check command
	deployCommand := deploy.Spec.Template.Spec.Containers[0].Command
	bootCommand := boot.Spec.Command
	if !reflect.DeepEqual(deployCommand, bootCommand) {
		logger.Info(reason, "type", "command", "Deploy", deploy.Name,
			"old", deployCommand, "new", bootCommand)
		deploy.Spec.Template.Spec.Containers[0].Command = bootCommand

		updated = true
	}

	if updated {
		err := c.Update(context.TODO(), deploy)
		if err != nil {
			logger.Info("Failed to update Deployment", "deploy", deploy.Name, "err", err.Error())
			handler.EventFail(reason, deploy.GetName(), err)

			return reconcile.Result{}, err, true
		}

		handler.EventNormal(reason, deploy.GetName())
		return reconcile.Result{Requeue: true}, nil, true
	}

	return reconcile.Result{}, nil, false
}

// reconcileUpdateService handle update logic of Service
func (handler *BootHandler) reconcileUpdateService(svc *corev1.Service) (reconcile.Result, error, bool) {
	boot := handler.Boot
	logger := handler.Logger
	c := handler.Client

	updated := false

	reason := "Updating Service"

	// 1. Check ownerReferences
	ownerReferences := svc.OwnerReferences
	if ownerReferences == nil || len(ownerReferences) == 0 {
		logger.Info(reason, "type", "ownerReferences", "service", svc.Name)

		_ = controllerutil.SetControllerReference(handler.OperatorBoot, svc, handler.Scheme)

		updated = true
	}

	// 2. Check port
	svcPort := svc.Spec.Ports[0].Port
	bootPort := boot.Spec.Port
	// Port changed
	if !(svcPort == bootPort) {
		logger.Info(reason, "type", "port", "service", svc.Name, "old", svcPort, "new", bootPort)
		svc.Spec.Ports[0].Port = bootPort
		svc.Spec.Ports[0].TargetPort = intstr.IntOrString{Type: intstr.Int, IntVal: int32(bootPort)}

		updated = true
	}

	// 3. Check annotation
	// Annotation is removed
	if svc.Annotations == nil {
		svc.Annotations = ServiceAnnotation(int(boot.Spec.Port))
		updated = true
	}

	// Annotation port is changed
	if svc.Annotations != nil {
		svcAnnoPort := svc.Annotations[PrometheusPortKey]
		if svcAnnoPort != strconv.Itoa(int(boot.Spec.Port)) {
			svc.Annotations[PrometheusPortKey] = strconv.Itoa(int(boot.Spec.Port))
			updated = true
		}
	}

	// 4. Check sessionAffinity
	svcAffinity := string(svc.Spec.SessionAffinity)
	bootAffinity := boot.Spec.SessionAffinity
	if !(svcAffinity == bootAffinity) {
		logger.Info(reason, "type", "sessionAffinity", "service", svc.Name, "old", svcAffinity, "new", bootAffinity)
		svc.Spec.SessionAffinity = corev1.ServiceAffinity(bootAffinity)

		updated = true
	}

	if updated {
		err := c.Update(context.TODO(), svc)
		if err != nil {
			logger.Error(err, "Failed to update Service", "type", "port", "service", svc.Name)
			handler.EventFail(reason, svc.GetName(), err)

			return reconcile.Result{}, err, true
		}

		handler.EventNormal(reason, svc.GetName())
	}

	return reconcile.Result{}, nil, false
}

// Refer https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
// CustomResourceSubresources(feature-gate) is only available in 1.10(Alpha), 1.11(Beta, default true)
func (handler *BootHandler) ReconcileUpdateBootMeta() (reconcile.Result, error, bool, bool) {
	logger := handler.Logger
	boot := handler.Boot
	c := handler.Client

	// 1. Update Deployment's metadata/annotations if needed
	depFound := &appsv1.Deployment{}
	depName := DeployName(boot)
	err := c.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: boot.Namespace}, depFound)
	if err != nil {
		logger.Error(err, "Failed to get Deployment")

		return reconcile.Result{}, err, true, false
	}

	// 2. Update Service's metadata/annotations if needed
	svcLabels := ServiceLabels(boot)
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{
		Namespace:     boot.Namespace,
		LabelSelector: labels.SelectorFromSet(svcLabels),
	}
	err = c.List(context.TODO(), listOptions, svcList)
	if err != nil {
		logger.Error(err, "Failed to list services")

		return reconcile.Result{}, err, true, false
	}

	annotationMap := map[string]string{
		DeployAnnotationKey:   depFound.Name,
		AppTypeAnnotationKey:  AppTypeAnnotationDeploy,
		ServicesAnnotationKey: TransferServiceNames(svcList.Items),
	}

	updated := handler.UpdateAnnotation(annotationMap)

	return reconcile.Result{}, nil, false, updated
}

func Ignore(namespace string) bool {
	oEnv := logan.OperDev

	isDevNamespace := strings.HasSuffix(namespace, "-dev")

	// "dev" namespace
	if oEnv == "dev" && !isDevNamespace {
		return true
	}

	// not "dev" namespace
	if oEnv != "dev" && isDevNamespace {
		return true
	}

	return false
}
