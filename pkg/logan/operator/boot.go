package operator

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HttpPortName is the Boot's created service port name
	HttpPortName = "http"

	// PrometheusPortKey is the Boot's created service's prometheus annotation key
	PrometheusPortKey = "prometheus.io/port"

	BootNameKey = "bootName"
)

// DeployLabels return labels for the created Deploy
func DeployLabels(boot *appv1.Boot) map[string]string {
	return map[string]string{"app": "havok", "havok/type": boot.Name}
}

// DeployName return name for the created Deploy
func DeployName(boot *appv1.Boot) string {
	return boot.Name
	//return boot.Name + "-" + boot.BootType
}

// AppContainerHealthPort return the health port for the created Pod's app container
func AppContainerHealthPort(boot *appv1.Boot, appSpec *config.AppSpec) intstr.IntOrString {
	healthPort := int32(boot.Spec.Port)
	if appSpec.Settings.AppHealthPort > 0 {
		healthPort = appSpec.Settings.AppHealthPort
	}
	return intstr.IntOrString{Type: intstr.Int, IntVal: int32(healthPort)}
}

// AppContainerImageName return image name for the created Pod's app container
func AppContainerImageName(boot *appv1.Boot, appSpec *config.AppSpec) string {
	registry := appSpec.Settings.Registry
	if registry != "" {
		return registry + "/" + boot.Spec.Image + ":" + boot.Spec.Version
	}

	return boot.Spec.Image + ":" + boot.Spec.Version
}

// PodLabels return labels for the created Pod
func PodLabels(boot *appv1.Boot) map[string]string {
	//return map[string]string{"app": "havok", boot.AppKey: boot.Name}
	return map[string]string{"app": "havok", BootNameKey: boot.Name, "bootType": boot.BootType}
}

// SideCarServiceName return the name for sidecar service
func SideCarServiceName(boot *appv1.Boot, port corev1.ContainerPort) string {
	return boot.Name + "-" + port.Name
}

// NodePortServiceName return the name for nodeport service
func NodePortServiceName(boot *appv1.Boot) string {
	return boot.Name + "-external"
}

// ServiceLabels return the labels for the created Service
func ServiceLabels(boot *appv1.Boot) map[string]string {
	return map[string]string{"app": boot.Name, "logan/env": logan.OperDev}
}

// ServiceAnnotation return the annotations for the created Service
func allowPrometheusScrape(boot *appv1.Boot, appSpec *config.AppSpec) bool {
	if boot.Spec.Prometheus != "" {
		prometheusScrape, _ := strconv.ParseBool(boot.Spec.Prometheus)
		return prometheusScrape
	}
	return *appSpec.Settings.PrometheusScrape
}

func ServiceAnnotation(prometheusScrape bool, port int) map[string]string {
	if prometheusScrape == true {
		return map[string]string{
			"prometheus.io/path":   "/prometheus",
			PrometheusPortKey:      strconv.Itoa(port),
			"prometheus.io/scheme": "http",
			"prometheus.io/scrape": "true",
		}
	}
	return nil
}

// TransferServiceNames transfer the services list of []Service to string, split by ,
func TransferServiceNames(services []corev1.Service) string {
	var serviceNames []string
	for _, service := range services {
		serviceNames = append(serviceNames, service.Name)
	}
	return strings.Join(serviceNames, ",")
}

// Decode will decode the origin string, with the fields of Boot
// Currently key only supports ${APP} and ${ENV} and ${PORT}
func Decode(boot *appv1.Boot, origin string) (string, bool) {
	ret := origin
	replaced := false
	if strings.Contains(origin, "${APP}") {
		ret = strings.ReplaceAll(origin, "${APP}", boot.Name)
		replaced = true
	}

	if strings.Contains(origin, "${ENV}") {
		ret = strings.ReplaceAll(ret, "${ENV}", logan.OperDev)
		replaced = true
	}
	if strings.Contains(origin, "${PORT}") {
		ret = strings.ReplaceAll(ret, "${PORT}",
			strconv.FormatInt(int64(boot.Spec.Port), 10))
		replaced = true
	}

	return ret, replaced
}

// DecodeEnvs replace the envVars, transforms the key with ${APP} and ${ENV} and ${PORT}
func DecodeEnvs(boot *appv1.Boot, envVars []corev1.EnvVar) bool {
	updated := false
	for i, envVar := range envVars {
		replaceEnv := envVar.DeepCopy()
		value, replaced := Decode(boot, envVar.Value)
		replaceEnv.Value = value
		if replaced {
			updated = true
		}
		envVars[i] = *replaceEnv
	}

	return updated
}

// DecodeVolumes replace the volumes, transforms the ClaimName with ${APP} and ${ENV} and ${PORT}
func DecodeVolumes(boot *appv1.Boot, volumes []corev1.Volume) bool {
	updated := false
	for i, volume := range volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		replaceVol := volume.DeepCopy()
		value, replaced := Decode(boot, volume.PersistentVolumeClaim.ClaimName)
		replaceVol.PersistentVolumeClaim.ClaimName = value
		if replaced {
			updated = true
		}
		volumes[i] = *replaceVol
	}
	return updated
}

// MarshalEnvVars marshal the []EnvVar to string
func MarshalEnvVars(envs []corev1.EnvVar) (string, error) {
	configEnvsSe, err := json.Marshal(envs)
	configEnvs := fmt.Sprintf("%s", configEnvsSe)
	return configEnvs, err
}

// DecodeEnvVars unmarshal the string to []EnvVar
func DecodeEnvVars(str string) ([]corev1.EnvVar, error) {
	var envVars []corev1.EnvVar

	err := json.Unmarshal([]byte(str), &envVars)

	return envVars, err
}

// EnvVarsEq return true if env1 and env2 is equal.
func EnvVarsEq(env1, env2 []corev1.EnvVar) bool {
	// If one is nil, the other must also be nil.
	if (env1 == nil) != (env2 == nil) {
		return false
	}

	if len(env1) != len(env2) {
		return false
	}

	for i := range env1 {
		aEnv := env1[i]
		bEnv := env2[i]
		if aEnv != bEnv {
			return false
		}
	}

	return true
}

// GetConfigSpec returns the config.AppSpec for the Boot.
func GetConfigSpec(boot *appv1.Boot) *config.AppSpec {
	if boot.BootType == logan.BootJava {
		return config.JavaConfig.AppSpec
	} else if boot.BootType == logan.BootPhp {
		return config.PhpConfig.AppSpec
	} else if boot.BootType == logan.BootPython {
		return config.PythonConfig.AppSpec
	} else if boot.BootType == logan.BootNodeJS {
		return config.NodeJSConfig.AppSpec
	} else if boot.BootType == logan.BootWeb {
		return config.WebConfig.AppSpec
	}

	return nil
}

// DecodeAnnotationEnvs decodes the annotation's env
func DecodeAnnotationEnvs(boot *appv1.Boot) ([]corev1.EnvVar, error) {
	bootMetaEnvsStr := boot.Annotations[BootEnvsAnnotationKey]
	if bootMetaEnvsStr == "" {
		// Boot's env is empty.
		return nil, nil
	}

	bootMetaEnvs, err := DecodeEnvVars(bootMetaEnvsStr)
	if err != nil {
		//Decoding error. Ignore validating.
		return nil, fmt.Errorf("decoding Annotation's env error. %s/%s: %s", boot.Namespace, boot.Name, err.Error())
	}

	return bootMetaEnvs, nil
}

// GetProfileBootConfig gets the Boot's config by profile annotation
func GetProfileBootConfig(boot *appv1.Boot, logger logr.Logger) (*config.BootConfig, error) {
	if boot.Annotations != nil {
		if _, exist := boot.Annotations[config.BootProfileAnnotationKey]; exist {
			bootProfile := boot.Annotations[config.BootProfileAnnotationKey]
			if bootProfile == logan.BootJava || bootProfile == logan.BootPhp || bootProfile == logan.BootPython || bootProfile == logan.BootNodeJS || bootProfile == logan.BootWeb {
				return nil, fmt.Errorf("boot using profile, but profile [%s] is not allow", bootProfile)
			}
			profileConfig := config.ProfileConfig[bootProfile]
			if profileConfig != nil {
				logger.Info("Boot using profile: ", "profile", bootProfile)
				return profileConfig, nil
			}
			return nil, fmt.Errorf("Boot using profile, but profile [%s] config is empty: ", bootProfile)
		}
	}

	return nil, nil
}

// GetCurrentTimestamp get the current time json string, as creationTimestamp of kubernetes
func GetCurrentTimestamp() string {
	now := metav1.Now()
	nowBytes, _ := now.MarshalJSON()
	return string(nowBytes[:])

}
