package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	logancfg "github.com/logancloud/logan-app-operator/pkg/logan/config"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/logancloud/logan-app-operator/pkg/apis"
	"github.com/logancloud/logan-app-operator/pkg/controller"

	mgrMetrics "github.com/logancloud/logan-app-operator/cmd/manager/metrics"
	mgrwebhook "github.com/logancloud/logan-app-operator/cmd/manager/webhook"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383

	configFile string
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))

	log.Info(fmt.Sprintf("Logan Operator Version: %s", logan.Version))
	log.Info(fmt.Sprintf("Logan Operator Inner Version: %s", logan.InnerVersion))
	log.Info(fmt.Sprintf("Logan Operator Env: %s", logan.OperDev))
	log.Info(fmt.Sprintf("Logan Operator Config: %s", configFile))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.StringVar(&configFile, "config", "configs/config.yaml", "The path to the logan operator config.")

	pflag.Parse()

	// Use a zap logr.Logger implementation.
	// If none of the zap flags are configured (or if the zap flag set is not being used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger implementing the logr.Logger interface.
	// This logger will be propagated through the whole operator, generating uniform and structured logs.

	logf.SetLogger(zap.Logger())
	//logf.SetLogger(logf.ZapLoggerTo(os.Stderr, true)) //Debug Output

	printVersion()

	err := logancfg.InitByFile(configFile)
	if err != nil {
		log.Error(err, "Init config file fail")
		os.Exit(1)
	}

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "logan-app-operator-lock"+"-"+logan.OperDev)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	metricsSvr, err := metrics.ExposeMetricsPort(ctx, metricsPort)
	if err != nil {
		log.Info(err.Error())
	}

	// add Prometheus Scrape
	_, err = mgrMetrics.AddPrometheusScrape(ctx, cfg, metricsSvr, metricsPort)
	if err != nil {
		log.Info(err.Error())
	}

	runningInCluster := true
	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		if err == k8sutil.ErrNoNamespace {
			runningInCluster = false
		}
	}

	if runningInCluster {
		mgrwebhook.RegisterWebhook(mgr, log, ns)
	} else {
		log.Info("Skipping registering webhook; not running in a cluster.")
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
