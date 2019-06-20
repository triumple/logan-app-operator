package logan

import (
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	// Version is the current operator version
	Version = "0.2.0"
	// Version is the current operator inner version
	InnerVersion = "1"
	defaultEnv   = "test"

	oEnvKey = "LOGAN_ENV"

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
}
