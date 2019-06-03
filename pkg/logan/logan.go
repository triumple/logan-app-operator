package logan

import (
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	Version      = "0.2.0"
	InnerVersion = "1"
	defaultEnv   = "test"

	OEnvKey = "LOGAN_ENV"

	BootJava   = "java"
	BootPhp    = "php"
	BootPython = "python"
	BootNodeJS = "nodejs"

	JavaAppKey   = "javaBoot"
	PhpAppKey    = "phpBoot"
	PythonAppKey = "pythonBoot"
	NodeJSAppKey = "nodejsBoot"
)

// OperDev is operator's running dev
var OperDev string

var log = logf.Log.WithName("logan_util")

func init() {
	ns, found := os.LookupEnv(OEnvKey)
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
}
