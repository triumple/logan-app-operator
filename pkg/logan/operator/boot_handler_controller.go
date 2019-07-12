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
// 1. Deployment not found: Create Deployment, requeue=true
// 2. Service not found: Create Service, requeue=true
// 3. When creating Error: requeue error requeue=true
func (handler *BootHandler) ReconcileCreate() (reconcile.Result, bool, error) {
	boot := handler.Boot
	logger := handler.Logger
	c := handler.Client
	requeue := false

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
				return reconcile.Result{}, true, err
			}

			handler.EventNormal(reason, dep.Name)
			depFound = dep
			requeue = true
		} else {
			logger.Error(err, "Failed to get Deployment")
			handler.EventFail(reason, depName, err)
			return reconcile.Result{}, true, err
		}
	}

	appSvcFound := &corev1.Service{}
	appSvcName := boot.Name
	err = c.Get(context.TODO(), types.NamespacedName{Name: appSvcName, Namespace: boot.Namespace}, appSvcFound)
	if err != nil && errors.IsNotFound(err) {
		reason := "Creating Service"
		if errors.IsNotFound(err) {
			// App Service not found, create app Service and sidecar services if necessary.
			// need to consider individually in the future

			// Creating all services
			for _, svc := range handler.NewServices(depFound) {
				logger.Info("Creating Service", "service", svc.Name)
				err = c.Create(context.TODO(), svc)
				if err != nil {
					//Note: Maybe when it called, the service is not created yet.
					logger.Info("Failed to create new Service, maybe the service is not created successfully",
						"service", svc.Name)
					handler.EventFail(reason, svc.Name, err)
					return reconcile.Result{}, true, nil
				}
				handler.EventNormal(reason, svc.Name)
				requeue = true
			}
		} else {
			logger.Error(err, "Failed to get Services")
			handler.EventFail(reason, boot.Name, err)
			return reconcile.Result{}, true, err
		}
	}
	// requeue the reconcile, if create deploy and service, k8s need sometime to create it.
	return reconcile.Result{Requeue: requeue}, requeue, nil
}

// ReconcileUpdate check the fields of components, if not as desire, update it.
// 1. Check Deployment's existence: error -> requeue=true
// 1.1. Check Deployment's fields: "replicas", image, env, port, resources, health, nodeSelector
// 2. Check Service's existence: error -> requeue=true
// 2.1 Check Service's fields:
func (handler *BootHandler) ReconcileUpdate() (reconcile.Result, bool, error) {
	boot := handler.Boot
	logger := handler.Logger
	c := handler.Client

	//1 Deployment
	depFound := &appsv1.Deployment{}
	depName := DeployName(boot)
	err := c.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: boot.Namespace}, depFound)
	if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return reconcile.Result{Requeue: true}, true, err
	}
	result, requeue, err := handler.reconcileUpdateDeploy(depFound)
	if requeue {
		return result, true, err
	}

	//2 Service
	appSvcFound := &corev1.Service{}
	appSvcName := boot.Name
	err = c.Get(context.TODO(), types.NamespacedName{Name: appSvcName, Namespace: boot.Namespace}, appSvcFound)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Service resource not found. Ignoring since object is not created successfully yet", "error", err)
			return reconcile.Result{Requeue: true}, true, nil
		}
		logger.Error(err, "Failed to get Service")
		return reconcile.Result{Requeue: true}, true, err
	}
	result, requeue, err = handler.reconcileUpdateService(appSvcFound, depFound)
	if requeue {
		return result, true, err
	}

	return reconcile.Result{}, false, nil
}

// reconcileUpdateDeploy handle update logic of Deployment
func (handler *BootHandler) reconcileUpdateDeploy(deploy *appsv1.Deployment) (reconcile.Result, bool, error) {
	logger := handler.Logger
	boot := handler.Boot
	c := handler.Client

	updated := false
	rebootUpdated := false

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

		rebootUpdated = true
	}

	// 4. Check env: check fist container(boot container)
	deployEnv := deploy.Spec.Template.Spec.Containers[0].Env
	bootEnv := boot.Spec.Env
	if !reflect.DeepEqual(deployEnv, bootEnv) {
		logger.Info(reason, "type", "env", "deploy", deploy.Name,
			"old", deployEnv, "new", bootEnv)

		rebootUpdated = true
	}

	// 5. Check port: check fist container(boot container)
	deployPorts := deploy.Spec.Template.Spec.Containers[0].Ports
	bootPorts := []corev1.ContainerPort{{Name: HttpPortName, ContainerPort: boot.Spec.Port, Protocol: corev1.ProtocolTCP}}
	if !reflect.DeepEqual(deployPorts, bootPorts) {
		logger.Info(reason, "type", "port", "deploy", deploy.Name,
			"old", deployPorts, "new", bootPorts)

		rebootUpdated = true
	}

	// 6 Check resources: check fist container(boot container)
	deployResources := deploy.Spec.Template.Spec.Containers[0].Resources
	bootResources := boot.Spec.Resources
	if !reflect.DeepEqual(deployResources, bootResources) {
		logger.Info(reason, "type", "resources", "deploy", deploy.Name,
			"old", deployResources, "new", bootResources)

		rebootUpdated = true
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

			rebootUpdated = true
		}
	} else {
		if probe == nil {
			// 1. If probe is nil, add Liveness and Readiness
			logger.Info(reason, "type", "health", "deploy", deploy.Name,
				"old", "empty", "new", bootHealth)

			rebootUpdated = true
		} else {
			deployHealth := probe.HTTPGet.Path
			// 2. If probe is not nil, we only need to update the health path
			if deployHealth != bootHealth {
				logger.Info(reason, "type", "health", "deploy", deploy.Name,
					"old", deployHealth, "new", bootHealth)

				rebootUpdated = true
			}
		}
	}

	// 8 Check nodeSelector: map[string]string
	deployNodeSelector := deploy.Spec.Template.Spec.NodeSelector
	bootNodeSelector := boot.Spec.NodeSelector
	if !reflect.DeepEqual(deployNodeSelector, bootNodeSelector) {
		logger.Info(reason, "type", "nodeSelector", "deploy", deploy.Name,
			"old", deployNodeSelector, "new", bootNodeSelector)

		rebootUpdated = true
	}

	// 9 Check command
	deployCommand := deploy.Spec.Template.Spec.Containers[0].Command
	bootCommand := boot.Spec.Command
	if !reflect.DeepEqual(deployCommand, bootCommand) {
		logger.Info(reason, "type", "command", "Deploy", deploy.Name,
			"old", deployCommand, "new", bootCommand)

		rebootUpdated = true
	}

	if rebootUpdated {
		updateDeploy := handler.NewDeployment()
		deploy.Spec = updateDeploy.Spec
	}

	if updated || rebootUpdated {
		err := c.Update(context.TODO(), deploy)
		if err != nil {
			logger.Info("Failed to update Deployment", "deploy", deploy.Name, "err", err.Error())
			handler.EventFail(reason, deploy.GetName(), err)

			return reconcile.Result{}, true, err
		}

		handler.EventNormal(reason, deploy.GetName())
		return reconcile.Result{Requeue: true}, true, nil
	}

	return reconcile.Result{}, false, nil
}

// reconcileUpdateService handle update logic/sidecar of Service
func (handler *BootHandler) reconcileUpdateService(svc *corev1.Service, deploy *appsv1.Deployment) (reconcile.Result, bool, error) {
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
	prometheusScrape := allowPrometheusScrape(boot, handler.Config.AppSpec)
	if svc.Annotations == nil && prometheusScrape == true {
		svc.Annotations = ServiceAnnotation(prometheusScrape, int(boot.Spec.Port))
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

	//PrometheusScrape is changed
	if svc.Annotations != nil && prometheusScrape == false {
		svc.Annotations = ServiceAnnotation(prometheusScrape, int(boot.Spec.Port))
		updated = true
	}

	// 4. Check sessionAffinity
	svcAffinity := string(svc.Spec.SessionAffinity)
	bootAffinity := boot.Spec.SessionAffinity
	if svcAffinity == "None" || svcAffinity == "" {
		//K8S will set sessionAffinity to "None" if the field is empty,
		if bootAffinity != "None" && bootAffinity != "" {
			logger.Info(reason, "type", "sessionAffinity", "service", svc.Name, "old", svcAffinity, "new", bootAffinity)
			svc.Spec.SessionAffinity = corev1.ServiceAffinity(bootAffinity)
			updated = true
		}
	} else {
		if !(svcAffinity == bootAffinity) {
			logger.Info(reason, "type", "sessionAffinity", "service", svc.Name, "old", svcAffinity, "new", bootAffinity)
			svc.Spec.SessionAffinity = corev1.ServiceAffinity(bootAffinity)
			updated = true
		}
	}

	if updated {
		err := c.Update(context.TODO(), svc)
		if err != nil {
			logger.Error(err, "Failed to update Service", "service", svc.Name)
			handler.EventFail(reason, svc.GetName(), err)

			return reconcile.Result{Requeue: true}, true, err
		}

		handler.EventNormal(reason, svc.GetName())
	}

	//handle update sidecar of Service
	result, requeue, err := handler.reconcileUpdateSidecarService(deploy)
	if err != nil {
		return reconcile.Result{Requeue: true}, true, err
	}

	if requeue {
		return result, requeue, err
	}

	return reconcile.Result{}, false, nil
}

// reconcileUpdateSidecarService handle update sidecar of Service
func (handler *BootHandler) reconcileUpdateSidecarService(deploy *appsv1.Deployment) (reconcile.Result, bool, error) {
	c := handler.Client
	boot := handler.Boot
	logger := handler.Logger

	runtimeSvcs, err := handler.listRuntimeService()
	if err != nil {
		return reconcile.Result{Requeue: true}, true, err
	}

	updated := false
	expectSvcs := handler.NewServices(deploy)
	// update or delete  sidecar service
	for _, runtimeSvc := range runtimeSvcs.Items {
		// skip app service
		if runtimeSvc.Name == boot.Name {
			continue
		}

		found := false
		modify := false
		for _, expectSvc := range expectSvcs {
			// skip app service
			if expectSvc.Name == boot.Name {
				continue
			}

			if runtimeSvc.Name == expectSvc.Name {
				found = true

				// 1. check ports
				// port\name
				runtimePort := runtimeSvc.Spec.Ports[0]
				expectPort := expectSvc.Spec.Ports[0]
				if runtimePort.Name != expectPort.Name ||
					runtimePort.Port != expectPort.Port {
					modify = true
					runtimeSvc.Spec.Ports = expectSvc.Spec.Ports
				}

				// 2. check OwnerReferences
				ownerReferences := runtimeSvc.OwnerReferences
				if ownerReferences == nil || len(ownerReferences) == 0 {
					runtimeSvc.OwnerReferences = expectSvc.OwnerReferences
				}

				// 3. check SessionAffinity
				if runtimeSvc.Spec.SessionAffinity != expectSvc.Spec.SessionAffinity {
					modify = true
					runtimeSvc.Spec.SessionAffinity = expectSvc.Spec.SessionAffinity
				}

				// 4. Check annotation
				// Annotation is removed
				if runtimeSvc.Annotations == nil {
					modify = true
					runtimeSvc.Annotations = expectSvc.Annotations
				}

				// Annotation port is changed
				if runtimeSvc.Annotations != nil {
					svcAnnoPort := runtimeSvc.Annotations[PrometheusPortKey]
					if svcAnnoPort != strconv.Itoa(int(expectSvc.Spec.Ports[0].Port)) {
						runtimeSvc.Annotations = expectSvc.Annotations
						modify = true
					}
				}
			}
		}

		if !found {
			logger.Info("Deleting Sidecar Service", "service", runtimeSvc.Name)

			err := c.Delete(context.TODO(), &runtimeSvc)
			if err != nil {
				logger.Error(err, "Failed to delete Sidecar Service", "service", runtimeSvc.Name)
				handler.EventFail("Failed to delete Sidecar Service", runtimeSvc.Name, err)
				return reconcile.Result{Requeue: true}, true, err
			}

			updated = true
		} else if modify {
			logger.Info("Updating Sidecar Service", "service", runtimeSvc.Name)

			err := c.Update(context.TODO(), &runtimeSvc)
			if err != nil {
				logger.Error(err, "Failed to update Sidecar Service", "service", runtimeSvc.Name)
				handler.EventFail("Failed to update Sidecar Service", runtimeSvc.Name, err)
				return reconcile.Result{Requeue: true}, true, err
			}

			updated = true
		}
	}

	// new sidecar service
	for _, expectSvc := range expectSvcs {
		// skip app service
		if expectSvc.Name == boot.Name {
			continue
		}

		notFound := true
		for _, runtimeSvc := range runtimeSvcs.Items {
			if runtimeSvc.Name == expectSvc.Name {
				notFound = false
			}
		}

		if notFound {
			logger.Info("Creating Sidecar Service", "service", expectSvc.Name)
			err := c.Create(context.TODO(), expectSvc)
			if err != nil {
				logger.Error(err, "Failed to create Sidecar Service", "service", expectSvc.Name)
				handler.EventFail("Failed to create Sidecar Service", expectSvc.Name, err)
				return reconcile.Result{Requeue: true}, true, err
			}
			updated = true
		}
	}

	if updated {
		return reconcile.Result{Requeue: true}, updated, nil
	}

	return reconcile.Result{}, false, nil
}

func (handler *BootHandler) listRuntimeService() (*corev1.ServiceList, error) {
	logger := handler.Logger
	boot := handler.Boot
	c := handler.Client
	svcLabels := ServiceLabels(boot)
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{
		Namespace:     boot.Namespace,
		LabelSelector: labels.SelectorFromSet(svcLabels),
	}
	err := c.List(context.TODO(), listOptions, svcList)
	if err != nil {
		logger.Error(err, "Failed to list services")
		return nil, err
	}
	return svcList, nil
}

// ReconcileUpdateBootMeta will handle the metadata update
// Refer https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
// CustomResourceSubresources(feature-gate) is only available in 1.10(Alpha), 1.11(Beta, default true)
func (handler *BootHandler) ReconcileUpdateBootMeta() (reconcile.Result, bool, bool, error) {
	logger := handler.Logger
	boot := handler.Boot
	c := handler.Client

	// 1. Update Deployment's metadata/annotations if needed
	depFound := &appsv1.Deployment{}
	depName := DeployName(boot)
	err := c.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: boot.Namespace}, depFound)
	if err != nil {
		logger.Error(err, "Failed to get Deployment")

		return reconcile.Result{}, true, false, err
	}

	// 2. Update Service's metadata/annotations if needed
	svcList, err := handler.listRuntimeService()
	if err != nil {
		logger.Error(err, "Failed to list services")

		return reconcile.Result{}, true, false, err
	}

	// 3. Update Boot's annotation if needed.
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(PodLabels(boot))
	listOptions := &client.ListOptions{Namespace: boot.Namespace, LabelSelector: labelSelector}
	err = c.List(context.TODO(), listOptions, podList)
	if err != nil {
		logger.Error(err, "Failed to list pods")
		return reconcile.Result{}, true, false, err
	}

	runningCount := len(podList.Items)
	//for _, pod := range podList.Items {
	//	podStatus := pod.Status.Phase
	//	if podStatus == corev1.PodRunning {
	//		runningCount = runningCount + 1
	//	}
	//}

	annotationMap := map[string]string{
		DeployAnnotationKey:          depFound.Name,
		AppTypeAnnotationKey:         AppTypeAnnotationDeploy,
		ServicesAnnotationKey:        TransferServiceNames(svcList.Items),
		StatusAvailableAnnotationKey: strconv.Itoa(runningCount),
		StatusDesiredAnnotationKey:   strconv.Itoa(int(*boot.Spec.Replicas)),
	}

	updated := handler.UpdateAnnotation(annotationMap)

	return reconcile.Result{}, false, updated, nil
}

// Ignore returns whether we should ignore handling for the Boot, decided by the Namespace
func Ignore(namespace string) bool {
	oEnv := logan.OperDev

	isDevNamespace := strings.HasSuffix(namespace, "-dev")
	isAutoNamespace := strings.HasSuffix(namespace, "-auto")

	// "dev" namespace
	if oEnv == "dev" && !isDevNamespace {
		return true
	}

	// "auto" namespace
	if oEnv == "auto" && !isAutoNamespace {
		return true
	}

	// normal namespaces
	if oEnv != "dev" && oEnv != "auto" {
		if isDevNamespace || isAutoNamespace {
			return true
		}
	}

	return false
}
