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
	operatorAppKey = "app"

	defaultPort    = 8080
	defaultReplica = 1
	defaultHealth  = "/health"

	// BootProfileAnnotationKey is the annotation key for storing boot's profile
	BootProfileAnnotationKey = "logan/profile"
)

var log = logf.Log.WithName("logan_config")

// BootConfig is the config struct for Boot.
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
	// JavaConfig is the config for JavaBoot
	JavaConfig *BootConfig
	// PhpConfig is the config for PhpBoot
	PhpConfig *BootConfig
	// PythonConfig is the config for PythonBoot
	PythonConfig *BootConfig
	// NodeJSConfig is the config for NodeJSBoot
	NodeJSConfig *BootConfig
	// WebConfig is the config for WebBoot
	WebConfig *BootConfig
	// ProfileConfig is the profile support config for All Boots, support to override the default profile.
	ProfileConfig map[string]*BootConfig
)

// SettingsConfig is the common struct for Settings
type SettingsConfig struct {
	Registry      string `json:"registry"`
	AppHealthPort int32  `json:"appHealthPort"`
}

// GlobalConfig is the entry for all boot's config
// 	- "java": default Java operator config
// 	- "php": default Php operator config
// 	- "python": default Python operator config
// 	- "nodejs": default NodeJS operator config
// 	- "web": default Web operator config
type GlobalConfig map[string]*OperatorConfig

// OperatorConfig is the struct for boot's global config
// 	- Operator's default settings：settings
// 	- operator's env specific config，oEnvs
//	- Container Info
//		1. application(app) container config: app
//		2. sidecar containers：sidecarContainers
//		3. sidecar services：sidecarServices
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

// InitByFile will initialize the config from the file
func InitByFile(configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		log.Error(err, "Can not open config file")
		os.Exit(1)
	}

	return Init(f)
}

// Init will initialize the config from the io.Reader
func Init(content io.Reader) error {
	err := NewConfig(content)
	if err != nil {
		return err
	}

	return nil
}

// NewConfig will initialize the config from the io.Reader
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

// NewConfigFromString will initialize the config from string, for testing,
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
		appSpec.Port = defaultPort
	}

	if appSpec.Replicas <= 0 {
		appSpec.Replicas = defaultReplica
	}

	if appSpec.Health == "" {
		appSpec.Health = defaultHealth
	}

	if appSpec.Settings == nil {
		appSpec.Settings = &SettingsConfig{}
	}

	// 1. Merge oEnv's settings: app
	appEnvDefault := operatorCfg.OEnvs[operatorAppKey][logan.OperDev]
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

// DecodeImageName will decode the image name from context
func DecodeImageName(image string, appSpec *AppSpec) string {
	registry := appSpec.Settings.Registry
	if registry != "" {
		return strings.ReplaceAll(image, "${REGISTRY}", registry)
	}

	return image
}
