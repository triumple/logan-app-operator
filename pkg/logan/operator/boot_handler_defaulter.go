package operator

import (
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	corev1 "k8s.io/api/core/v1"
)

// Setting the default value for CR
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) DefaultValue() bool {
	logger := handler.Logger
	appConfigSpec := handler.Config.AppSpec
	bootSpec := handler.OperatorSpec

	changed := false

	if bootSpec == nil {
		logger.Info("Defaulters", "bootSpec is nil")
		return false
	}

	//port
	if bootSpec.Port <= 0 {
		logger.Info("Defaulters", "type", "port", "spec", bootSpec.Port, "default", appConfigSpec.Port)
		bootSpec.Port = appConfigSpec.Port
		changed = true
	}

	//replicas
	if bootSpec.Replicas == nil || *bootSpec.Replicas < 0 {
		// User did not specify value
		logger.Info("Defaulters", "type", "replicas", "spec", nil, "default", appConfigSpec.Replicas)
		newReplicas := appConfigSpec.Replicas
		bootSpec.Replicas = &newReplicas
		changed = true
	}

	//health: "/health"
	if bootSpec.Health == "" {
		logger.Info("Defaulters", "type", "health", "spec", bootSpec.Health, "default", appConfigSpec.Health)
		bootSpec.Health = appConfigSpec.Health
		changed = true
	}

	//resources.limits
	if len(bootSpec.Resources.Limits) == 0 {
		resourcesList := appConfigSpec.Resources.Limits
		if len(resourcesList) > 0 {
			logger.Info("Defaulters", "type", "resources.limits", "spec", bootSpec.Resources.Limits,
				"default", resourcesList)
			bootSpec.Resources.Limits = resourcesList
			changed = true
		}
	}

	//resources.request
	if len(bootSpec.Resources.Requests) == 0 {
		resourcesList := appConfigSpec.Resources.Requests
		if len(resourcesList) > 0 {
			logger.Info("Defaulters", "type", "resources.requests", "spec", bootSpec.Resources.Requests,
				"default", resourcesList)
			bootSpec.Resources.Requests = resourcesList
			changed = true
		}
	}

	//check cpu limits and request:  cpu request>limit, set request=limit
	requestCpu := bootSpec.Resources.Requests.Cpu()
	limitCpu := bootSpec.Resources.Limits.Cpu()
	if requestCpu.Cmp(*limitCpu) > 0 {
		logger.Info("Defaulters", "type", "request.cpu", "spec", requestCpu,
			"default", limitCpu)

		bootSpec.Resources.Requests[corev1.ResourceCPU] = *limitCpu
		changed = true
	}

	//check mem limits and request: mem request>limit, set request=limit
	requestMem := bootSpec.Resources.Requests.Memory()
	limitMem := bootSpec.Resources.Limits.Memory()
	if requestMem.Cmp(*limitMem) > 0 {
		logger.Info("Defaulters", "type", "request.mem.request", "spec", requestMem,
			"default", limitMem)

		bootSpec.Resources.Requests[corev1.ResourceMemory] = *limitMem

		changed = true
	}

	//subDomain:
	if bootSpec.SubDomain == "" {
		logger.Info("Defaulters", "type", "subDomain", "spec", bootSpec.SubDomain, "default", appConfigSpec.SubDomain)
		bootSpec.SubDomain = appConfigSpec.SubDomain
		changed = true
	}

	//nodeSelector: ""
	if len(bootSpec.NodeSelector) == 0 {
		selector := appConfigSpec.NodeSelector
		if selector != nil && len(selector) > 0 {
			logger.Info("Defaulters", "type", "nodeSelector", "spec", bootSpec.NodeSelector, "default", selector)
			bootSpec.NodeSelector = selector
			changed = true
		}
	} else {
		// Boot is specified the nodeSelector, should check config and merge.
		nodeSelectorMerge := false
		for key, value := range appConfigSpec.NodeSelector {
			bootValue := bootSpec.NodeSelector[key]
			if bootValue != "" {
				if bootValue != value {
					bootSpec.NodeSelector[key] = value
					nodeSelectorMerge = true
				}
			} else {
				bootSpec.NodeSelector[key] = value
				nodeSelectorMerge = true
			}
		}

		if nodeSelectorMerge {
			logger.Info("Defaulters", "type", "nodeSelector", "to", bootSpec.NodeSelector)
			changed = true
		}
	}

	envChanged := handler.DefaultEnvValue()

	return changed || envChanged
}

// DefaultEnvValue will handle the env changed.
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) DefaultEnvValue() bool {
	logger := handler.Logger
	boot := handler.Boot
	appConfigSpec := handler.Config.AppSpec
	bootSpec := handler.OperatorSpec
	bootMeta := handler.OperatorMeta

	changed := false

	//env:
	// annotation-1: Boot Spec is modified，clear value of annotation's env and generation
	if bootMeta.Annotations == nil {
		bootMeta.Annotations = make(map[string]string)
	}
	bootMetaEnvsStr := bootMeta.Annotations[BootEnvsAnnotationKey]
	if bootMetaEnvsStr == "" {
		// Annotation "boot-envs" is empty, means it is newly created.
		// Clear value of annotation's env to let it do the Env Defaulters.
		bootMeta.Annotations[EnvAnnotationKey] = ""
	} else {
		// Annotation "boot-envs" is not empty, we need to check if it is modified.
		bootMetaEnvs, err := DecodeEnvVars(bootMetaEnvsStr)
		if err != nil {
			logger.Error(err, "Decoding annotation's env error.")
			return false
		}

		// Check the annotation's env and Boot's env
		if EnvVarsEq(bootMetaEnvs, bootSpec.Env) {
			if handler.ImageChange() {
				// Image is changed, we need to merge the envs.
				bootMeta.Annotations[EnvAnnotationKey] = ""
			} else {
				// Not changed, do nothing.
				return false
			}
		} else {
			// Env is changed, clear value of annotation's env to let it do the Env Defaulters.
			bootMeta.Annotations[EnvAnnotationKey] = ""
		}
	}

	metaEnv := bootMeta.Annotations[EnvAnnotationKey]
	// annotation-2: New Boot or Boot's spec is modified, do the Env Defaulters.
	if metaEnv != "" {
		return false
	}

	annotationMap := map[string]string{
		EnvAnnotationKey:        string(EnvAnnotationValue),
		BootImagesAnnotationKey: AppContainerImageName(handler.Boot, handler.Config.AppSpec),
	}

	updated := handler.UpdateAnnotation(annotationMap)
	if updated {
		mergeEnvs := make([]corev1.EnvVar, 0)
		for _, specEnv := range appConfigSpec.Env {
			env := specEnv.DeepCopy()
			mergeEnvs = append(mergeEnvs, *env)
		}
		DecodeEnvs(boot, mergeEnvs)

		added := appv1.BootSpec{
			Env: mergeEnvs,
		}

		logger.Info("Defaulters", "type", "env", "spec", bootSpec.Env, "default", added.Env)

		err := util.MergeOverride(bootSpec, added)
		if err != nil {
			logger.Error(err, "config merge error.", "type", "default")
		}

		DecodeEnvs(boot, bootSpec.Env)

		changed = true

		logger.Info("Defaulters", "init env changed", changed)
	} else {
		// In case user could change the env after created. we need to check the env value
		changed = DecodeEnvs(boot, bootSpec.Env)

		logger.Info("Defaulters", "user env changed", changed)
	}

	// If changed, Save Boot's envs into annotation, to make it compare latter.
	bootEnvStr, err := MarshalEnvVars(bootSpec.Env)
	if err != nil {
		logger.Error(err, "Encoding boot'env error.")
	}
	bootMeta.Annotations[BootEnvsAnnotationKey] = bootEnvStr

	return changed
}

// ImageChange will handle the image changed.
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) ImageChange() bool {
	bootMeta := handler.OperatorMeta

	if bootMeta.Annotations == nil {
		bootMeta.Annotations = make(map[string]string)
	}
	bootMetaImageStr := bootMeta.Annotations[BootImagesAnnotationKey]
	if bootMetaImageStr == "" {
		// Annotation "boot-images" is empty, means it is newly created.
		return true
	} else {
		// Annotation "boot-images" is not empty, we need to check if it is modified.
		return bootMetaImageStr != AppContainerImageName(handler.Boot, handler.Config.AppSpec)
	}
}
