package metrics

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

const (
	// PrometheusPathAnnotationKey is the Boot's created service's prometheus path annotation key
	PrometheusPathAnnotationKey = "prometheus.io/path"
	// PrometheusPathAnnotationValue is the Boot's created service's prometheus path annotation value
	PrometheusPathAnnotationValue = "/metrics"

	// PrometheusPortAnnotationKey is the Boot's created service's prometheus port annotation key
	PrometheusPortAnnotationKey = "prometheus.io/port"

	// PrometheusSchemeAnnotationKey is the Boot's created service's prometheus scheme annotation key
	PrometheusSchemeAnnotationKey = "prometheus.io/scheme"
	// PrometheusSchemeAnnotationValue is the Boot's created service's prometheus scheme annotation value
	PrometheusSchemeAnnotationValue = "http"

	// PrometheusScrapeAnnotationKey is the Boot's created service's prometheus scrape annotation key
	PrometheusScrapeAnnotationKey = "prometheus.io/scrape"
	// PrometheusScrapeAnnotationValue is the Boot's created service's prometheus scrape annotation value
	PrometheusScrapeAnnotationValue = "true"
)

var log = logf.Log.WithName("metrics")

// AddPrometheusScrape will add prometheus annotations to service
func AddPrometheusScrape(ctx context.Context, config *rest.Config, svr *v1.Service, port int32) (*v1.Service, error) {
	if svr.Annotations == nil {
		svr.Annotations = make(map[string]string)
	}

	svr.Annotations[PrometheusPathAnnotationKey] = PrometheusPathAnnotationValue
	svr.Annotations[PrometheusPortAnnotationKey] = strconv.Itoa(int(port))
	svr.Annotations[PrometheusSchemeAnnotationKey] = PrometheusSchemeAnnotationValue
	svr.Annotations[PrometheusScrapeAnnotationKey] = PrometheusScrapeAnnotationValue

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
