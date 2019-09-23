package metrics

import (
	"context"
	"fmt"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

var log = logf.Log.WithName("metrics")

// AddPrometheusScrape will add prometheus annotations to service
func AddPrometheusScrape(ctx context.Context, config *rest.Config, svr *v1.Service, port int32) (*v1.Service, error) {
	if svr.Annotations == nil {
		svr.Annotations = make(map[string]string)
	}

	svr.Annotations[keys.PrometheusPathAnnotationKey] = keys.PrometheusPathAnnotationValue
	svr.Annotations[keys.PrometheusPortAnnotationKey] = strconv.Itoa(int(port))
	svr.Annotations[keys.PrometheusSchemeAnnotationKey] = keys.PrometheusSchemeAnnotationValue
	svr.Annotations[keys.PrometheusScrapeAnnotationKey] = keys.PrometheusScrapeAnnotationValue

	service, err := updatePrometheusService(ctx, config, svr)
	if err != nil {
		return nil, fmt.Errorf("failed to create or get service for metrics with prometheus scrape: %v", err)
	}

	return service, nil
}

func updatePrometheusService(ctx context.Context, config *rest.Config, svr *v1.Service) (*v1.Service, error) {
	client, err := createClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %v", err)
	}
	service, err := updateService(ctx, client, svr)
	if err != nil {
		return nil, fmt.Errorf("failed to create or get service for metrics with prometheus scrape: %v", err)
	}

	return service, nil
}

func createClient(config *rest.Config) (crclient.Client, error) {
	client, err := crclient.New(config, crclient.Options{})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// from https://github.com/operator-framework/operator-sdk/blob/master/pkg/metrics/metrics.go  createOrUpdateService
func updateService(ctx context.Context, client crclient.Client, s *v1.Service) (*v1.Service, error) {
	existingService := &v1.Service{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: s.Namespace,
	}, existingService)
	if err != nil {
		return nil, err
	}

	s.ResourceVersion = existingService.ResourceVersion
	if existingService.Spec.Type == v1.ServiceTypeClusterIP {
		s.Spec.ClusterIP = existingService.Spec.ClusterIP
	}
	err = client.Update(ctx, s)
	if err != nil {
		return nil, err
	}
	log.Info("Metrics Service object updated", "Service.Name", s.Name, "Service.Namespace", s.Namespace)
	return s, nil
}
