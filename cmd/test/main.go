package main

import (
	"fmt"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func printObj(cfg *config.BootConfig, bootType string) {
	fmt.Println(bootType)
	fmt.Println("	spec: ", cfg.AppSpec)
	fmt.Println("	spec env: ", cfg.AppSpec.Env)
	fmt.Println("	spec env len: ", len(cfg.AppSpec.Env))
	fmt.Println("	spec container: ", cfg.AppSpec.Container)
	if cfg.AppSpec.Container != nil {
		fmt.Println("	spec container lifecycle: ", cfg.AppSpec.Container.Lifecycle)
		fmt.Println("	spec container volumeMounts: ", cfg.AppSpec.Container.VolumeMounts)
	}
	if cfg.AppSpec.PodSpec != nil {
		fmt.Println("	spec podSpec Volumes: ", cfg.AppSpec.PodSpec.Volumes)
		fmt.Println("	spec podSpec initContainers: ", cfg.AppSpec.PodSpec.InitContainers)
	}

	fmt.Println("	sidecar containers: ", cfg.SidecarContainers)
	if cfg.SidecarContainers != nil {
		for _, c := range *cfg.SidecarContainers {
			fmt.Println("	sidecar container: ", c)
			fmt.Println("	  sidecar container name: ", c.Name)
			fmt.Println("	  sidecar container env: ", c.Env)
		}
	}

	fmt.Println("	sidecar services: ", cfg.SidecarServices)
	fmt.Println("	registry: ", cfg.AppSpec.Settings.Registry)
	fmt.Println("	subDomain: ", cfg.AppSpec.SubDomain)
	fmt.Println("	appHealthPort: ", cfg.AppSpec.Settings.AppHealthPort)
	fmt.Println("	registry: ", cfg.AppSpec.Settings.Registry)
	fmt.Println("	request.memory: ", cfg.AppSpec.Resources.Requests.Memory())
	fmt.Println("	request.cpu: ", cfg.AppSpec.Resources.Requests.Cpu())
	fmt.Println("	limits.memory: ", cfg.AppSpec.Resources.Limits.Memory())
	fmt.Println("	limits.cpu: ", cfg.AppSpec.Resources.Limits.Cpu())
	fmt.Println("")
}

func main() {
	logan.OperDev = "dev"
	logf.SetLogger(logf.ZapLoggerTo(os.Stderr, true)) //Debug Output

	file := "logan-app-operator/configs/config.yaml"
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	err = config.Init(f)
	if err != nil {
		panic(err)
	}

	printObj(config.JavaConfig, "java")
	printObj(config.PhpConfig, "php")
	printObj(config.PythonConfig, "python")
	printObj(config.NodeJSConfig, "nodejs")
}
