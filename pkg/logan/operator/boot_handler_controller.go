package operator

import (
	"context"
	"fmt"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	loganMetrics "github.com/logancloud/logan-app-operator/pkg/logan/metrics"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
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
		if errors.IsNotFound(err) {
			dep := handler.NewDeployment()
			logger.Info("Creating Deployment", "deploy containers", dep.Spec.Template.Spec.Containers)
			err = c.Create(context.TODO(), dep)
			if err != nil {
				msg := fmt.Sprintf("Failed to create Deployment: %s", depName)
				logger.Error(err, msg)
				loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_CREATE_STAGE, loganMetrics.RECONCILE_CREATE_DEPLOYMENT_SUBSTAGE, boot.Name)
				handler.RecordEvent(keys.FailedCreateDeployment, msg, err)
				return reconcile.Result{}, true, err
			}

			handler.RecordEvent(keys.CreatedDeployment, fmt.Sprintf("Created Deployment: %s", depName), nil)
			depFound = dep
			requeue = true
		} else {
			msg := fmt.Sprintf("Failed to get Deployment: %s", depName)
			logger.Error(err, msg)
			loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_CREATE_STAGE, loganMetrics.RECONCILE_GET_DEPLOYMENT_SUBSTAGE, boot.Name)
			handler.RecordEvent(keys.FailedGetDeployment, msg, err)
			return reconcile.Result{}, true, err
		}
	}

	appSvcFound := &corev1.Service{}
	appSvcName := boot.Name
	err = c.Get(context.TODO(), types.NamespacedName{Name: appSvcName, Namespace: boot.Namespace}, appSvcFound)
	if err != nil && errors.IsNotFound(err) {
		if errors.IsNotFound(err) {
			// App Service not found, create app Service and sidecar/nodePort services if necessary.
			// need to consider individually in the future

			// Creating all services
			for _, svc := range handler.NewServices(depFound) {
				logger.Info("Creating Service", "service", svc.Name)
				err = c.Create(context.TODO(), svc)
				if err != nil {
					//Note: Maybe when it called, the service is not created yet.
					msg := fmt.Sprintf("Failed to create Service: %s", svc.Name)
					logger.Error(err, msg)
					loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_CREATE_STAGE, loganMetrics.RECONCILE_CREATE_SERVICE_SUBSTAGE, boot.Name)
					handler.RecordEvent(keys.FailedCreateService, msg, err)
					return reconcile.Result{}, true, nil
				}
				handler.RecordEvent(keys.CreatedService, fmt.Sprintf("Created Service: %s", svc.Name), nil)
				requeue = true
			}
		} else {
			msg := fmt.Sprintf("Failed to get Service: %s", appSvcName)
			logger.Error(err, msg)
			loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_CREATE_STAGE, loganMetrics.RECONCILE_GET_SERVICE_SUBSTAGE, boot.Name)
			handler.RecordEvent(keys.FailedGetService, msg, err)
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
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_GET_DEPLOYMENT_SUBSTAGE, boot.Name)
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
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_GET_SERVICE_SUBSTAGE, boot.Name)
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

	// 10 Check vol
	deployVols := deploy.Spec.Template.Spec.Containers[0].VolumeMounts
	bootVolStr, ok := boot.Annotations[keys.BootDeployPvcsAnnotationKey]
	if ok && bootVolStr != "" {
		bootVols, err := DecodeVolumeMountVars(bootVolStr)
		if err != nil {
			logger.Error(err, "can not decode VolumeMount", "Deploy", deploy.Name,
				keys.BootDeployPvcsAnnotationKey, bootVolStr)
			return reconcile.Result{Requeue: true}, true, err
		}

		if !VolumeMountVarsEq(deployVols, bootVols) {
			deleted, added, modified := util.DifferenceVol(deployVols, bootVols)
			logger.Info("Boot VolumeMounts change.", "Deploy", deploy.Name,
				"deleted", deleted, "added", added, "modified", modified)

			volUpdated, err := handler.checkVolumeMountUpdate(deleted, added, modified)
			if err != nil {
				logger.Error(err, "Fail to reconcile VolumeMounts", "Deploy", deploy.Name,
					"deleted", deleted, "added", added, "modified", modified)
				return reconcile.Result{Requeue: true}, true, err
			}

			if volUpdated {
				logger.Info(reason, "type", "VolumeMounts", "Deploy", deploy.Name,
					"old", deployVols, "new", bootVols, keys.BootDeployPvcsAnnotationKey, bootVolStr)
				rebootUpdated = true
			}
		}
	} else if deployVols != nil {
		// if we have this, is very surprised
		logger.Info(reason, "type", "VolumeMounts", "Deploy", deploy.Name,
			"old", deployVols, "new", nil)
		// rebootUpdated = true
	}

	if rebootUpdated {
		updateDeploy := handler.NewDeployment()
		deploy.Spec = updateDeploy.Spec
		logger.Info("this update will cause rolling update", "Deploy", deploy.Name)
	}

	if updated || rebootUpdated {
		err := c.Update(context.TODO(), deploy)
		if err != nil {
			msg := fmt.Sprintf("Failed to update Deployment: %s", deploy.GetName())
			logger.Info(msg, "err", err.Error())
			loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_UPDATE_DEPLOYMENT_SUBSTAGE, boot.Name)
			handler.RecordEvent(keys.FailedUpdateDeployment, msg, err)

			return reconcile.Result{Requeue: true}, true, err
		}

		handler.RecordEvent(keys.UpdatedDeployment, fmt.Sprintf("Updated Deployment: %s", deploy.GetName()), nil)
		return reconcile.Result{Requeue: true}, true, nil
	}

	return reconcile.Result{}, false, nil
}

// reconcileUpdateService handle update logic/sidecar/nodePort of Service
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
		svcAnnoPort := svc.Annotations[keys.PrometheusPortAnnotationKey]
		if svcAnnoPort != strconv.Itoa(int(boot.Spec.Port)) {
			svc.Annotations[keys.PrometheusPortAnnotationKey] = strconv.Itoa(int(boot.Spec.Port))
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
			msg := fmt.Sprintf("Failed to update Service: %s", svc.GetName())
			logger.Error(err, msg)
			loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_UPDATE_SERVICE_SUBSTAGE, boot.Name)
			handler.RecordEvent(keys.FailedUpdateService, msg, err)

			return reconcile.Result{Requeue: true}, true, err
		}

		handler.RecordEvent(keys.UpdatedService, fmt.Sprintf("Updated Service: %s", svc.GetName()), nil)
	}

	//handle update sidecar/nodePort of Service
	result, requeue, err := handler.reconcileUpdateOtherService(deploy)
	if err != nil {
		return reconcile.Result{Requeue: true}, true, err
	}

	if requeue {
		return result, requeue, err
	}

	return reconcile.Result{}, false, nil
}

func (handler *BootHandler) checkVolumeMountUpdate(deleted, added, modified []corev1.VolumeMount) (bool, error) {
	c := handler.Client
	boot := handler.Boot

	checkPvc := func(vols []corev1.VolumeMount) (bool, error) {
		if vols != nil {
			for _, vol := range vols {
				pvc := &corev1.PersistentVolumeClaim{}
				err := c.Get(context.TODO(), types.NamespacedName{Namespace: boot.Namespace, Name: vol.Name}, pvc)
				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return false, err
				}

				if pvc.Labels != nil {
					// is shared pvc
					shared, found := pvc.Labels[keys.SharedKey]
					if found {
						if "true" == shared {
							return true, nil
						}
					}

					// is boot private pvc
					podLabels := PodLabels(boot)
					if reflect.DeepEqual(podLabels, pvc.Labels) {
						return true, nil
					}
				}
			}
		}
		return false, nil
	}

	allVols := make([]corev1.VolumeMount, 0)
	allVols = append(allVols, deleted...)
	allVols = append(allVols, added...)
	allVols = append(allVols, modified...)

	volUpdated, err := checkPvc(allVols)
	return volUpdated, err
}

// reconcileUpdateOtherService handle update sidecar/nodePort of Service
func (handler *BootHandler) reconcileUpdateOtherService(deploy *appsv1.Deployment) (reconcile.Result, bool, error) {
	c := handler.Client
	boot := handler.Boot
	logger := handler.Logger

	runtimeSvcs, err := handler.listRuntimeService()
	if err != nil {
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_LIST_SERVICES_SUBSTAGE, boot.Name)
		return reconcile.Result{Requeue: true}, true, err
	}

	updated := false
	expectSvcs := handler.NewServices(deploy)
	// update or delete  sidecar/nodePort service
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
					!reflect.DeepEqual(runtimePort.TargetPort, expectPort.TargetPort) {
					modify = true
					runtimeSvc.Spec.Ports = expectSvc.Spec.Ports
				}

				if runtimeSvc.Spec.Type == corev1.ServiceTypeClusterIP &&
					runtimePort.Port != expectPort.Port {
					modify = true
					runtimeSvc.Spec.Ports[0].Port = expectPort.Port
				}

				// make sure Port equal to NodePort
				if runtimeSvc.Spec.Type == corev1.ServiceTypeNodePort &&
					runtimePort.NodePort != runtimePort.Port {
					modify = true
					runtimeSvc.Spec.Ports[0].Port = runtimePort.NodePort
					runtimeSvc.Spec.Ports[0].NodePort = runtimePort.NodePort
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
				if runtimeSvc.Annotations == nil && expectSvc.Annotations != nil {
					modify = true
					runtimeSvc.Annotations = expectSvc.Annotations
				}

				// Annotation port is changed
				if runtimeSvc.Annotations != nil {
					svcAnnoPort := runtimeSvc.Annotations[keys.PrometheusPortAnnotationKey]
					if svcAnnoPort != strconv.Itoa(int(expectSvc.Spec.Ports[0].Port)) {
						runtimeSvc.Annotations = expectSvc.Annotations
						modify = true
					}
				}
			}
		}

		if !found {
			logger.Info("Deleting Other Service", "service", runtimeSvc.Name)

			err := c.Delete(context.TODO(), &runtimeSvc)
			if err != nil {
				msg := fmt.Sprintf("Failed to delete Other Service: %s", runtimeSvc.Name)
				logger.Error(err, msg)
				loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_DELETE_OTHER_SERVICE_SUBSTAGE, boot.Name)
				handler.RecordEvent(keys.FailedDeleteService, msg, err)
				return reconcile.Result{Requeue: true}, true, err
			}
			handler.RecordEvent(keys.DeletedService, fmt.Sprintf("Deleted Other Service: %s", runtimeSvc.Name), nil)

			updated = true
		} else if modify {
			logger.Info("Updating Other Service", "service", runtimeSvc.Name)

			err := c.Update(context.TODO(), &runtimeSvc)
			if err != nil {
				msg := fmt.Sprintf("Failed to update Other Service: %s", runtimeSvc.Name)
				logger.Error(err, msg)
				loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_UPDATE_OTHER_SERVICE_SUBSTAGE, boot.Name)
				handler.RecordEvent(keys.FailedUpdateService, msg, err)
				return reconcile.Result{Requeue: true}, true, err
			}
			handler.RecordEvent(keys.UpdatedService, fmt.Sprintf("Updated Other Service: %s", runtimeSvc.Name), nil)

			updated = true
		}
	}

	// new sidecar/nodePort service
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
			logger.Info("Creating Other Service", "service", expectSvc.Name)
			err := c.Create(context.TODO(), expectSvc)
			if err != nil {
				msg := fmt.Sprintf("Failed to create Other Service: %s", expectSvc.Name)
				logger.Error(err, msg)
				loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_STAGE, loganMetrics.RECONCILE_CREATE_OTHER_SERVICE_SUBSTAGE, boot.Name)
				handler.RecordEvent(keys.FailedCreateService, msg, err)
				return reconcile.Result{Requeue: true}, true, err
			}
			handler.RecordEvent(keys.CreatedService, fmt.Sprintf("Created Other Service: %s", expectSvc.Name), nil)

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
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_BOOT_META_STAGE, loganMetrics.RECONCILE_GET_DEPLOYMENT_SUBSTAGE, boot.Name)
		return reconcile.Result{}, true, false, err
	}

	// 2. Update Service's metadata/annotations if needed
	svcList, err := handler.listRuntimeService()
	if err != nil {
		logger.Error(err, "Failed to list services")
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_BOOT_META_STAGE, loganMetrics.RECONCILE_LIST_SERVICES_SUBSTAGE, boot.Name)
		return reconcile.Result{}, true, false, err
	}

	// 3. Update Boot's annotation if needed.
	// 3.1 Update Boot's annotation StatusAvailable
	podList := &corev1.PodList{}
	podLabels := PodLabels(boot)
	labelSelector := labels.SelectorFromSet(podLabels)
	listOptions := &client.ListOptions{Namespace: boot.Namespace, LabelSelector: labelSelector}
	err = c.List(context.TODO(), listOptions, podList)
	if err != nil {
		logger.Error(err, "Failed to list pods")
		loganMetrics.UpdateReconcileErrors(boot.Kind, loganMetrics.RECONCILE_UPDATE_BOOT_META_STAGE, loganMetrics.RECONCILE_LIST_PODS_SUBSTAGE, boot.Name)
		return reconcile.Result{}, true, false, err
	}

	runningCount := depFound.Status.Replicas
	//for _, pod := range podList.Items {
	//	podStatus := pod.Status.Phase
	//	if podStatus == corev1.PodRunning {
	//		runningCount = runningCount + 1
	//	}
	//}

	// 3.2 Update Boot's annotation revision
	//   select latest revision. set it
	revisionLst, _ := c.ListRevision(boot.Namespace, podLabels)
	latestRevision := revisionLst.SelectLatestRevision()

	// 3.2.1 Update Boot's revison's annotation
	//    set the latest revison's phase to active
	revisionAnnotationMap := map[string]string{}
	if runningCount == *boot.Spec.Replicas && runningCount == depFound.Status.AvailableReplicas {
		revisionAnnotationMap[keys.BootRevisionPhaseAnnotationKey] = RevisionPhaseActive
	} else {
		revisionAnnotationMap[keys.BootRevisionPhaseAnnotationKey] = RevisionPhaseRunning
	}

	if latestRevision != nil {
		revisionUpdated := updateRevisionAnnotation(latestRevision, revisionAnnotationMap)
		if revisionUpdated {
			reason := "Updating Boot Revision Meta"
			logger.Info(reason, "new", revisionAnnotationMap, "revision", latestRevision)
			err := c.Update(context.TODO(), latestRevision)
			if err != nil {
				msg := "Failed to update Boot Revision Meta"
				logger.Info(msg, "err", err.Error())
				handler.RecordEvent(keys.FailedUpdateBootMeta, msg, err)
				return reconcile.Result{Requeue: true}, true, false, err
			}
			handler.RecordEvent(keys.UpdatedBootMeta, "Updated Boot Revision Meta", nil)
		}
	} else {
		logger.Info("can not find latest Revision", "boot", boot)
	}

	//requeue := false
	//if revisionAnnotationMap[keys.BootRevisionPhaseAnnotationKey] == RevisionPhaseRunning {
	//	logger.V(1).Info("Revision is running,should requeue", "revision", latestRevision)
	//	requeue = true
	//}

	// 4 Update boot Meta
	annotationMap := map[string]string{
		keys.DeployAnnotationKey:          depFound.Name,
		keys.AppTypeAnnotationKey:         keys.AppTypeAnnotationDeploy,
		keys.ServicesAnnotationKey:        TransferServiceNames(svcList.Items),
		keys.StatusAvailableAnnotationKey: strconv.Itoa(int(runningCount)),
		keys.StatusDesiredAnnotationKey:   strconv.Itoa(int(*boot.Spec.Replicas)),
	}

	if latestRevision != nil {
		annotationMap[keys.BootRevisionIdAnnotationKey] = strconv.Itoa(latestRevision.GetRevisionId())
	}

	updated := handler.UpdateAnnotation(annotationMap)

	//if requeue {
	//	return reconcile.Result{RequeueAfter: time.Second * 10}, true, updated, nil
	//}

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
