package config

import (
	"bytes"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"io"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

const (
	OperatorAppKey = "app"

	DefaultPort    = 8080
	DefaultReplica = 1
	DefaultHealth  = "/health"

	BootProfileAnnotationKey = "logan/profile"
)

var log = logf.Log.WithName("logan_config")

type BootConfig struct {
	AppSpec           *AppSpec
	SidecarContainers *[]corev1.Container
	SidecarServices   *[]SidecarService
}

// AppSpec define the App spec
type AppSpec struct {
	Type         string                  `json:"type"`
	Port         int32                   `json:"port"`
	Replicas     int32                   `json:"replicas"`
	Health       string                  `json:"health"`
	Env          []corev1.EnvVar         `json:"env"`
	Resources    v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string       `json:"nodeSelector"`
	SubDomain    string                  `json:"subDomain"`

	PodSpec   *corev1.PodSpec   `json:"podSpec"`
	Container *corev1.Container `json:"container"`
	Settings  *SettingsConfig   `json:"settings"`
}

// SidecarService define the service for Sidecar
type SidecarService struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
}

var (
	JavaConfig    *BootConfig
	PhpConfig     *BootConfig
	PythonConfig  *BootConfig
	NodeJSConfig  *BootConfig
	WebConfig     *BootConfig
	ProfileConfig map[string]*BootConfig
)

// 配置信息
type SettingsConfig struct {
	Registry      string `json:"registry"`
	AppHealthPort int32  `json:"appHealthPort"`
}

// 	"java": default Java operator config
// 	"php": default Php operator config
// 	"python": default Python operator config
// 	"nodejs": default NodeJS operator config
// 	"web": default Web operator config
type GlobalConfig map[string]*OperatorConfig

// 配置包含：
// 	Operator默认配置信息：settings
// 	operator的各个环境默认信息，oEnvs
//	容器信息
//		1. 应用app容器配置信息: app
//		2. sidecar容器信息：sidecarContainers
//		3. sidecar服务信息：sidecarServices
type OperatorConfig struct {
	// Operator配置信息
	Settings *SettingsConfig `json:"settings"`

	// Operator的环境信息配置
	OEnvs map[string]map[string]AppSpec `json:"oEnvs"`

	// Operator的默认App配置
	AppSpec *AppSpec `json:"app"`

	//Operator的默认SidecarContainers配置
	SidecarContainers *[]corev1.Container `json:"sideCarContainers"`

	// Sidecar的Service列表
	SidecarServices *[]SidecarService `json:"sidecarServices"`
}

func InitByFile(configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		log.Error(err, "Can not open config file")
		os.Exit(1)
	}

	return Init(f)
}

func Init(content io.Reader) error {
	err := NewConfig(content)
	if err != nil {
		return err
	}

	return nil
}

func NewConfig(content io.Reader) error {
	c := GlobalConfig{}

	err := k8syaml.NewYAMLOrJSONDecoder(content, 100).Decode(&c)
	if err != nil {
		return err
	}

	gConfig := c
	gConfig.applyDefaults()

	ProfileConfig = make(map[string]*BootConfig, 0)

	operator := gConfig[logan.BootJava]
	JavaConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig[logan.BootPhp]
	PhpConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig[logan.BootPython]
	PythonConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig[logan.BootNodeJS]
	NodeJSConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig[logan.BootWeb]
	WebConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	for key, operator := range gConfig {
		if key != logan.BootJava && key != logan.BootPhp && key != logan.BootPython && key != logan.BootNodeJS && key != logan.BootWeb {
			ProfileConfig[key] = &BootConfig{
				AppSpec: operator.AppSpec,

				SidecarContainers: operator.SidecarContainers,
				SidecarServices:   operator.SidecarServices,
			}
		}
	}

	return nil
}

func NewConfigFromString(content string) error {
	if content == "" {
		globalCfg := &GlobalConfig{}
		globalCfg.applyDefaults()
		return nil
	}

	return NewConfig(bytes.NewBuffer([]byte(content)))
}

func (globalCfg GlobalConfig) applyDefaults() {
	applyDefaultWithSidecar(globalCfg, globalCfg[logan.BootJava], logan.BootJava)

	applyDefaultWithSidecar(globalCfg, globalCfg[logan.BootPhp], logan.BootPhp)

	applyDefaultWithSidecar(globalCfg, globalCfg[logan.BootPython], logan.BootPython)

	applyDefaultWithSidecar(globalCfg, globalCfg[logan.BootNodeJS], logan.BootNodeJS)

	applyDefaultWithSidecar(globalCfg, globalCfg[logan.BootWeb], logan.BootWeb)

	for key, value := range globalCfg {
		if key != logan.BootJava && key != logan.BootPhp && key != logan.BootPython && key != logan.BootNodeJS && key != logan.BootWeb {
			applyDefaultWithSidecar(globalCfg, value, key)
		}
	}
}

func applyDefaultWithSidecar(globalCfg GlobalConfig, operatorCfg *OperatorConfig, bootType string) {
	if operatorCfg == nil {
		operatorCfg = &OperatorConfig{}
		if bootType == logan.BootJava {
			globalCfg[logan.BootJava] = operatorCfg
		} else if bootType == logan.BootPhp {
			globalCfg[logan.BootPhp] = operatorCfg
		} else if bootType == logan.BootPython {
			globalCfg[logan.BootPython] = operatorCfg
		} else if bootType == logan.BootNodeJS {
			globalCfg[logan.BootNodeJS] = operatorCfg
		} else if bootType == logan.BootWeb {
			globalCfg[logan.BootWeb] = operatorCfg
		} else {
			// Other profiles: use key
			globalCfg[bootType] = operatorCfg
		}
	}

	if operatorCfg.AppSpec == nil {
		operatorCfg.AppSpec = &AppSpec{}
	}
	appSpec := operatorCfg.AppSpec
	applyDefault(operatorCfg, appSpec, bootType)

	// Replace Registry's name
	// 1. InitContainers
	podSpec := operatorCfg.AppSpec.PodSpec
	if podSpec != nil && podSpec.InitContainers != nil {
		for i, container := range podSpec.InitContainers {
			podSpec.InitContainers[i].Image = DecodeImageName(container.Image, operatorCfg.AppSpec)
		}
	}

	// 2. SidecarContainers
	sidecarContainers := operatorCfg.SidecarContainers
	if sidecarContainers != nil {
		for i, container := range *sidecarContainers {
			//Replace Image Registry
			(*sidecarContainers)[i].Image = DecodeImageName(container.Image, operatorCfg.AppSpec)

			// Merge oEnv's settings: sidecar
			appEnvDefault := operatorCfg.OEnvs[container.Name][logan.OperDev]
			toMerge := AppSpec{
				Env: container.Env,
			}
			err := util.MergeOverride(&toMerge, appEnvDefault)
			(*sidecarContainers)[i].Env = toMerge.Env
			if err != nil {
				log.Error(err, "sidecar env config merge error.", "type", bootType)
			}
		}
	}
}

func applyDefault(operatorCfg *OperatorConfig, appSpec *AppSpec, bootType string) {
	if appSpec.Port <= 0 {
		appSpec.Port = DefaultPort
	}

	if appSpec.Replicas <= 0 {
		appSpec.Replicas = DefaultReplica
	}

	if appSpec.Health == "" {
		appSpec.Health = DefaultHealth
	}

	//if appSpec.Resources == nil {
	//	appSpec.Resources = &v1.ResourceRequirements{}
	//}

	if appSpec.Settings == nil {
		appSpec.Settings = &SettingsConfig{}
	}

	// 1. Merge oEnv's settings: app
	appEnvDefault := operatorCfg.OEnvs[OperatorAppKey][logan.OperDev]
	err := util.MergeOverride(appSpec, appEnvDefault)
	if err != nil {
		log.Error(err, "env config merge error.", "type", bootType)
	}

	// 2. Merge Operator's Settings
	// 2.1 Envs -> Global Settings
	oSettings := operatorCfg.Settings
	envSettings := appEnvDefault.Settings
	if envSettings != nil && oSettings != nil {
		if envSettings.Registry != "" {
			oSettings.Registry = envSettings.Registry
		}
		if envSettings.AppHealthPort > 0 {
			oSettings.AppHealthPort = envSettings.AppHealthPort
		}
	}

	// 2.2 Global Settings-> App Settings
	if oSettings != nil {
		// No use Mergo to avoid the whole struct override.
		if oSettings.Registry != "" {
			appSpec.Settings.Registry = oSettings.Registry
		}

		if oSettings.AppHealthPort > 0 {
			appSpec.Settings.AppHealthPort = oSettings.AppHealthPort
		}
	}
}

func DecodeImageName(image string, appSpec *AppSpec) string {
	registry := appSpec.Settings.Registry
	if registry != "" {
		return strings.ReplaceAll(image, "${REGISTRY}", registry)
	}

	return image
}
