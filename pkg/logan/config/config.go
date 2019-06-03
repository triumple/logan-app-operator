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
)

var log = logf.Log.WithName("logan_config")

type BootConfig struct {
	AppSpec           *AppSpec
	SidecarContainers *[]corev1.Container
	SidecarServices   *[]SidecarService
}

// AppSpec define the App spec
type AppSpec struct {
	Type         string                   `json:"type"`
	Port         int32                    `json:"port"`
	Replicas     int32                    `json:"replicas"`
	Health       string                   `json:"health"`
	Env          []corev1.EnvVar          `json:"env"`
	Resources    *v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string        `json:"nodeSelector"`
	SubDomain    string                   `json:"subDomain"`

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
	JavaConfig   *BootConfig
	PhpConfig    *BootConfig
	PythonConfig *BootConfig
	NodeJSConfig *BootConfig
)

// 配置信息
type SettingsConfig struct {
	Registry      string `json:"registry"`
	AppHealthPort int32  `json:"appHealthPort"`
}

// 配置包含：
// 	Java operator配置信息
// 	Php operator配置信息
// 	Python operator配置信息
// 	NodeJS operator配置信息
type globalConfig struct {
	// Java配置
	JavaOperatorConfig *operatorConfig `json:"java"`
	// Php配置
	PhpOperatorConfig *operatorConfig `json:"php"`
	// Python配置
	PythonOperatorConfig *operatorConfig `json:"python"`
	// NodeJS配置
	NodeJSOperatorConfig *operatorConfig `json:"nodejs"`
}

// 配置包含：
// 	Operator默认配置信息：settings
// 	operator的各个环境默认信息，oEnvs
//	容器信息
//		1. 应用app容器配置信息: app
//		2. sidecar容器信息：sidecarContainers
//		3. sidecar服务信息：sidecarServices
type operatorConfig struct {
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
	c := globalConfig{}

	err := k8syaml.NewYAMLOrJSONDecoder(content, 100).Decode(&c)
	if err != nil {
		return err
	}

	gConfig := &c
	gConfig.applyDefaults()

	operator := gConfig.JavaOperatorConfig
	JavaConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig.PhpOperatorConfig
	PhpConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig.PythonOperatorConfig
	PythonConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	operator = gConfig.NodeJSOperatorConfig
	NodeJSConfig = &BootConfig{
		AppSpec: operator.AppSpec,

		SidecarContainers: operator.SidecarContainers,
		SidecarServices:   operator.SidecarServices,
	}

	return nil
}

func NewConfigFromString(content string) error {
	if content == "" {
		globalCfg := &globalConfig{}
		globalCfg.applyDefaults()
		return nil
	}

	return NewConfig(bytes.NewBuffer([]byte(content)))
}

func (globalCfg *globalConfig) applyDefaults() {
	applyDefaultWithSidecar(globalCfg, globalCfg.JavaOperatorConfig, logan.BootJava)

	applyDefaultWithSidecar(globalCfg, globalCfg.PhpOperatorConfig, logan.BootPhp)

	applyDefaultWithSidecar(globalCfg, globalCfg.PythonOperatorConfig, logan.BootPython)

	applyDefaultWithSidecar(globalCfg, globalCfg.NodeJSOperatorConfig, logan.BootNodeJS)
}

func applyDefaultWithSidecar(globalCfg *globalConfig, operatorCfg *operatorConfig, bootType string) {
	if operatorCfg == nil {
		operatorCfg = &operatorConfig{}
		if bootType == logan.BootJava {
			globalCfg.JavaOperatorConfig = operatorCfg
		} else if bootType == logan.BootPhp {
			globalCfg.PhpOperatorConfig = operatorCfg
		} else if bootType == logan.BootPython {
			globalCfg.PythonOperatorConfig = operatorCfg
		} else if bootType == logan.BootNodeJS {
			globalCfg.NodeJSOperatorConfig = operatorCfg
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

func applyDefault(operatorCfg *operatorConfig, appSpec *AppSpec, bootType string) {
	if appSpec.Port <= 0 {
		appSpec.Port = DefaultPort
	}

	if appSpec.Replicas <= 0 {
		appSpec.Replicas = DefaultReplica
	}

	if appSpec.Health == "" {
		appSpec.Health = DefaultHealth
	}

	if appSpec.Resources == nil {
		appSpec.Resources = &v1.ResourceRequirements{}
	}

	if appSpec.Settings == nil {
		appSpec.Settings = &SettingsConfig{}
	}

	//1. Merge Operator's Settings
	oSettings := operatorCfg.Settings
	if oSettings != nil {
		// No use Mergo to avoid the whole struct override.
		if oSettings.Registry != "" {
			appSpec.Settings.Registry = oSettings.Registry
		}

		if oSettings.AppHealthPort > 0 {
			appSpec.Settings.AppHealthPort = oSettings.AppHealthPort
		}
	}

	// 2. Merge oEnv's settings: app
	appEnvDefault := operatorCfg.OEnvs[OperatorAppKey][logan.OperDev]
	err := util.MergeOverride(appSpec, appEnvDefault)
	if err != nil {
		log.Error(err, "env config merge error.", "type", bootType)
	}
}

func DecodeImageName(image string, appSpec *AppSpec) string {
	registry := appSpec.Settings.Registry
	if registry != "" {
		return strings.ReplaceAll(image, "${REGISTRY}", registry)
	}

	return image
}
