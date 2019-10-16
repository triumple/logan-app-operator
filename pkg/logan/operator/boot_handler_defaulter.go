package operator

import (
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

// DefaultValue will set the default value for CR
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
	if bootSpec.Port <= 0 && appConfigSpec.Port > 0 {
		logger.Info("Defaulters", "type", "port", "spec", bootSpec.Port, "default", appConfigSpec.Port)
		bootSpec.Port = appConfigSpec.Port
		changed = true
	}

	//replicas
	if bootSpec.Replicas == nil || *bootSpec.Replicas < 0 {
		// User did not specify value
		defaultReplicas := appConfigSpec.Replicas
		if defaultReplicas <= 0 {
			defaultReplicas = 1
		}
		logger.Info("Defaulters", "type", "replicas", "spec", nil, "default", defaultReplicas)
		bootSpec.Replicas = &defaultReplicas
		changed = true
	}

	//health
	if bootSpec.Health == nil && appConfigSpec.Health != "" {
		logger.Info("Defaulters", "type", "health", "spec", bootSpec.Health, "default", appConfigSpec.Health)
		overrideHealth := appConfigSpec.Health
		bootSpec.Health = &overrideHealth
		changed = true
	}

	//Prometheus
	if bootSpec.Prometheus == "" && appConfigSpec.Settings.PrometheusScrape != nil {
		logger.Info("Defaulters", "type", "prometheus", "spec", bootSpec.Prometheus, "default", appConfigSpec.Settings.PrometheusScrape)
		bootSpec.Prometheus = strconv.FormatBool(*appConfigSpec.Settings.PrometheusScrape)
		changed = true
	}

	//resources.limits
	if len(appConfigSpec.Resources.Limits) > 0 {
		limitResources := appConfigSpec.Resources.Limits.DeepCopy()
		if len(bootSpec.Resources.Limits) == 0 {
			logger.Info("Defaulters", "type", "resources.limits", "spec", bootSpec.Resources.Limits,
				"default", limitResources)
			bootSpec.Resources.Limits = limitResources
			changed = true
		} else {
			// boot's limits mem is empty
			bootSpecMem := bootSpec.Resources.Limits.Memory()
			if bootSpecMem.CmpInt64(0) == 0 {
				logger.Info("Defaulters", "type", "resources.limits.memory",
					"spec", bootSpec.Resources.Limits.Memory(),
					"default", bootSpecMem)
				bootSpec.Resources.Limits[corev1.ResourceMemory] = *limitResources.Memory()
				changed = true
			}

			// boot's limits cpu is empty
			bootSpecCpu := bootSpec.Resources.Limits.Cpu()
			if bootSpecCpu.CmpInt64(0) == 0 {
				logger.Info("Defaulters", "type", "resources.limits.cpu",
					"spec", bootSpec.Resources.Limits.Cpu(),
					"default", bootSpecCpu)
				bootSpec.Resources.Limits[corev1.ResourceCPU] = *limitResources.Cpu()
				changed = true
			}
		}
	}

	//resources.requests
	if len(appConfigSpec.Resources.Requests) > 0 {
		requestResources := appConfigSpec.Resources.Requests.DeepCopy()
		if len(bootSpec.Resources.Requests) == 0 {
			logger.Info("Defaulters", "type", "resources.requests", "spec", bootSpec.Resources.Requests,
				"default", requestResources)
			bootSpec.Resources.Requests = requestResources
			changed = true
		} else {
			// boot's requests mem is empty
			bootSpecMem := bootSpec.Resources.Requests.Memory()
			if bootSpecMem.CmpInt64(0) == 0 {
				logger.Info("Defaulters", "type", "resources.requests.memory",
					"spec", bootSpec.Resources.Requests.Memory(),
					"default", bootSpecMem)
				bootSpec.Resources.Requests[corev1.ResourceMemory] = *requestResources.Memory()
				changed = true
			}

			// boot's requests cpu is empty
			bootSpecCpu := bootSpec.Resources.Requests.Cpu()
			if bootSpecCpu.CmpInt64(0) == 0 {
				logger.Info("Defaulters", "type", "resources.requests.cpu",
					"spec", bootSpec.Resources.Requests.Cpu(),
					"default", bootSpecCpu)
				bootSpec.Resources.Requests[corev1.ResourceCPU] = *requestResources.Cpu()
				changed = true
			}
		}
	}

	//check cpu limits and request:  cpu request>limit, set request=limit
	requestCpu := bootSpec.Resources.Requests.Cpu()
	limitCpu := bootSpec.Resources.Limits.Cpu()
	if limitCpu.CmpInt64(0) != 0 && requestCpu.Cmp(*limitCpu) > 0 {
		logger.Info("Defaulters", "type", "request.cpu", "spec", requestCpu,
			"default", limitCpu)

		bootSpec.Resources.Requests[corev1.ResourceCPU] = *limitCpu
		changed = true
	}

	//check mem limits and request: mem request>limit, set request=limit
	requestMem := bootSpec.Resources.Requests.Memory()
	limitMem := bootSpec.Resources.Limits.Memory()
	if limitMem.CmpInt64(0) != 0 && requestMem.Cmp(*limitMem) > 0 {
		logger.Info("Defaulters", "type", "request.mem.request", "spec", requestMem,
			"default", limitMem)

		bootSpec.Resources.Requests[corev1.ResourceMemory] = *limitMem

		changed = true
	}

	//subDomain:
	if bootSpec.SubDomain == "" && appConfigSpec.SubDomain != "" {
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

	pvcChanged := handler.DefaultPvcValue()

	return changed || envChanged || pvcChanged
}

// DefaultPvcValue will handle the pvc changed.
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) DefaultPvcValue() bool {
	logger := handler.Logger
	boot := handler.Boot
	bootSpec := handler.OperatorSpec
	bootMeta := handler.OperatorMeta

	//Pvc:
	// annotation-1: Boot Spec is modified，clear value of annotation's pvc and generation
	if bootMeta.Annotations == nil {
		bootMeta.Annotations = make(map[string]string)
	}

	annotationMap := map[string]string{
		keys.BootPvcsAnnotationKey:       "",
		keys.BootDeployPvcsAnnotationKey: "",
	}

	updatePvcMeta := func() {
		if bootSpec.Pvc == nil || len(bootSpec.Pvc) == 0 {
			annotationMap[keys.BootPvcsAnnotationKey] = ""
		} else {
			pvcStr, err := MarshalPvcVars(bootSpec.Pvc)
			if err != nil {
				logger.Error(err, "Encoding boot's pvc error.")
			}
			annotationMap[keys.BootPvcsAnnotationKey] = pvcStr
		}

		vols := make([]corev1.VolumeMount, 0)
		if handler.Config.AppSpec.Container != nil &&
			handler.Config.AppSpec.Container.VolumeMounts != nil {
			vols = append(vols, handler.Config.AppSpec.Container.VolumeMounts...)
		}

		if bootSpec.Pvc != nil && len(bootSpec.Pvc) > 0 {
			vols = append(vols, ConvertVolumeMount(bootSpec.Pvc)...)
		}

		if len(vols) > 0 {
			DecodeVolumeMounts(boot, vols)
			volStr, err := MarshalVolumeMountVars(vols)
			if err != nil {
				logger.Error(err, "Encoding boot's vol error.")
			}
			annotationMap[keys.BootDeployPvcsAnnotationKey] = volStr
		} else {
			annotationMap[keys.BootDeployPvcsAnnotationKey] = ""
		}
	}

	bootMetaPvcStr, ok := bootMeta.Annotations[keys.BootPvcsAnnotationKey]
	// created boot or previous pvc is empty
	if !ok || bootMetaPvcStr == "" {
		updatePvcMeta()
	} else {
		//check update
		previousPvc, err := DecodePvcVars(bootMetaPvcStr)
		if err != nil {
			logger.Error(err, "Decoding annotation's pvc error.")
			return false
		}
		if PvcVarsEq(previousPvc, bootSpec.Pvc) {
			return false
		}
		updatePvcMeta()
	}

	updated := handler.UpdateAnnotation(annotationMap)
	return updated
}

// DefaultEnvValue will handle the env changed.
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) DefaultEnvValue() bool {
	logger := handler.Logger
	boot := handler.Boot
	appConfigSpec := handler.Config.AppSpec
	bootSpec := handler.OperatorSpec
	bootMeta := handler.OperatorMeta

	//env:
	// annotation-1: Boot Spec is modified，clear value of annotation's env and generation
	if bootMeta.Annotations == nil {
		bootMeta.Annotations = make(map[string]string)
	}
	bootMetaEnvsStr := bootMeta.Annotations[keys.BootEnvsAnnotationKey]
	if bootMetaEnvsStr == "" {
		// Annotation "boot-envs" is empty, means it is newly created.
		// Clear value of annotation's env to let it do the Env Defaulters.
		bootMeta.Annotations[keys.EnvAnnotationKey] = ""
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
				bootMeta.Annotations[keys.EnvAnnotationKey] = ""
			} else {
				// Not changed, do nothing.
				return false
			}
		} else {
			// Env is changed, clear value of annotation's env to let it do the Env Defaulters.
			bootMeta.Annotations[keys.EnvAnnotationKey] = ""
		}
	}

	metaEnv := bootMeta.Annotations[keys.EnvAnnotationKey]
	// annotation-2: New Boot or Boot's spec is modified, do the Env Defaulters.
	if metaEnv != "" {
		return false
	}

	annotationMap := map[string]string{
		keys.EnvAnnotationKey:        string(keys.EnvAnnotationValue),
		keys.BootImagesAnnotationKey: AppContainerImageName(handler.Boot, handler.Config.AppSpec),
	}

	var changed bool
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
	bootMeta.Annotations[keys.BootEnvsAnnotationKey] = bootEnvStr

	return changed
}

// ImageChange will handle the image changed.
// Return true if should be updated, false if should not be updated
func (handler *BootHandler) ImageChange() bool {
	bootMeta := handler.OperatorMeta

	if bootMeta.Annotations == nil {
		bootMeta.Annotations = make(map[string]string)
	}
	bootMetaImageStr := bootMeta.Annotations[keys.BootImagesAnnotationKey]
	if bootMetaImageStr == "" {
		// Annotation "boot-images" is empty, means it is newly created.
		return true
	}

	// Annotation "boot-images" is not empty, we need to check if it is modified.
	return bootMetaImageStr != AppContainerImageName(handler.Boot, handler.Config.AppSpec)
}
