package logan

import (
	"os"
	"runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
)

const (
	defaultEnv = "test"
	oEnvKey    = "LOGAN_ENV"

	defaultConfigMap = "logan-app-operator-config"
	oConfigMapKey    = "CONFIGMAP_NAME"

	// ConfigFilename is for config file name
	ConfigFilename = "config.yaml"

	oMutationDefaulterKey  = "MUTATION_DEFAULTER"
	oRevisionMaxHistoryKey = "MAX_HISTORY"
	oBizENVKey             = "BIZ_ENVS"

	// BootJava is for JavaBoot type
	BootJava = "java"
	// BootPhp is for PhpBoot type
	BootPhp = "php"
	// BootPython is for PythonBoot type
	BootPython = "python"
	// BootNodeJS is for NodeJSBoot type
	BootNodeJS = "nodejs"
	// BootWeb is for WebBoot type
	BootWeb = "web"

	// JavaAppKey is for JavaBoot type
	JavaAppKey = "javaBoot"
	// PhpAppKey is for PhpBoot type
	PhpAppKey = "phpBoot"
	// PythonAppKey is for PythonBoot type
	PythonAppKey = "pythonBoot"
	// NodeJSAppKey is for NodeJSBoot type
	NodeJSAppKey = "nodejsBoot"
	// WebAppKey is for WebBoot type
	WebAppKey = "webBoot"
)

// OperDev is operator's running dev
var OperDev string

// OperConfigmap is operator's config map
var OperConfigmap string

// MaxConcurrentReconciles is max concurrent reconciles
var MaxConcurrentReconciles int

// MutationDefaulter is whether to modify in the webhook phase
var MutationDefaulter bool

// MutationDefaulter is the maximum number of revisions retained
var MaxHistory int

// BizEnvs is what ENV needs to be filtered
var BizEnvs map[string]bool

var log = logf.Log.WithName("logan_util")

func init() {
	ns, found := os.LookupEnv(oEnvKey)
	if !found {
		log.Info("Env not set, use default", "env", defaultEnv)
		OperDev = defaultEnv
	}

	if ns == "" {
		log.Info("Env set is empty, use default", "env", defaultEnv)
		OperDev = defaultEnv
	} else {
		OperDev = ns
	}

	configMap, found := os.LookupEnv(oConfigMapKey)
	if !found {
		log.Info("CONFIGMAP_NAME not set, use default", "CONFIGMAP_NAME", defaultConfigMap)
		OperConfigmap = defaultConfigMap
	}

	if configMap == "" {
		log.Info("CONFIGMAP_NAME set is empty, use default", "CONFIGMAP_NAME", defaultConfigMap)
		OperConfigmap = defaultConfigMap
	} else {
		OperConfigmap = configMap
	}

	mutationDefaulter, found := os.LookupEnv(oMutationDefaulterKey)
	if !found {
		log.Info("MUTATION_DEFAULTER not set, use default", "MUTATION_DEFAULTER", false)
		MutationDefaulter = false
	} else {
		b, err := strconv.ParseBool(mutationDefaulter)
		if err != nil {
			log.Error(err, "MUTATION_DEFAULTER parse error, use default", "MUTATION_DEFAULTER", false)
			MutationDefaulter = false
		} else {
			MutationDefaulter = b
		}
	}

	maxHistory, found := os.LookupEnv(oRevisionMaxHistoryKey)
	if !found {
		log.Info("MAX_HISTORY not set, use default", "MAX_HISTORY", 10)
		MaxHistory = 10
	} else {
		i, err := strconv.Atoi(maxHistory)
		if err != nil {
			log.Error(err, "MAX_HISTORY parse error, use default", "MAX_HISTORY", 10)
			MaxHistory = 10
		} else {
			MaxHistory = i
		}
	}

	bizEnvs, found := os.LookupEnv(oBizENVKey)
	if !found {
		log.Info("BIZ_ENVS not set, use default", "BIZ_ENVS", "")
		BizEnvs = make(map[string]bool)
	} else {
		BizEnvs = make(map[string]bool)
		for _, val := range strings.Split(bizEnvs, ",") {
			BizEnvs[val] = true
		}
	}

	MaxConcurrentReconciles = runtime.NumCPU() * 2
}
