package main

import (
	"context"
	"fmt"
	logancfg "github.com/logancloud/logan-app-operator/pkg/logan/config"
	"k8s.io/apimachinery/pkg/types"
	"os"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	corev1 "k8s.io/api/core/v1"
)

var log = logf.Log.WithName("leader")

func main() {
	logf.SetLogger(logf.ZapLoggerTo(os.Stderr, true)) //Debug Output
	log.Info("Reading configmaps")

	config, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Get config")
		panic(err)
	}

	client, err := crclient.New(config, crclient.Options{})

	cmap := &corev1.ConfigMap{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: "logan",
		Name:      "logan-app-operator-config",
	}, cmap)

	if err != nil {
		log.Error(err, "get configmap")
		panic(err)
	}

	fmt.Println("uid", cmap.UID)

	configContent, found := cmap.Data["config.yaml"]

	if !found {
		panic("Cluster Monitoring ConfigMap does not contain a config. Using defaults.")
	}

	err = logancfg.NewConfigFromString(configContent)
	if err != nil {
		log.Error(err, "Cluster Monitoring config could not be parsed. Using defaults: %v")
		panic(err)
	}

	fmt.Println("Java App Spec", logancfg.JavaConfig.AppSpec)
}
