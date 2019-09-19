package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"time"
)

const (
	// RECONCILE_GET_BOOT_STAGE is main stage to get boot from store
	// Following stages are main stages
	RECONCILE_GET_BOOT_STAGE = "reconcile_get_boot"

	// RECONCILE_UPDATE_BOOT_DEFAULTERS_STAGE is main stage to update boot with defaulters
	RECONCILE_UPDATE_BOOT_DEFAULTERS_STAGE = "reconcile_update_boot_defaulters"

	// RECONCILE_CREATE_STAGE is main stage to create deployment, service, etc.
	RECONCILE_CREATE_STAGE = "reconcile_create"

	// RECONCILE_UPDATE_STAGE is main stage to update deployment, service, etc.
	RECONCILE_UPDATE_STAGE = "reconcile_update"

	// RECONCILE_UPDATE_BOOT_META_STAGE is main stage to update boot .gmetadata.
	RECONCILE_UPDATE_BOOT_META_STAGE = "reconcile_update_boot_meta"

	// RECONCILE_CREATE_DEPLOYMENT_SUBSTAGE is sub stage to create deployment.
	// Following stages are sub stages
	RECONCILE_CREATE_DEPLOYMENT_SUBSTAGE = "create_deployment"

	// RECONCILE_GET_DEPLOYMENT_SUBSTAGE is sub stage to get deployment.
	RECONCILE_GET_DEPLOYMENT_SUBSTAGE = "get_deployment"

	// RECONCILE_UPDATE_DEPLOYMENT_SUBSTAGE is sub stage to update deployment.
	RECONCILE_UPDATE_DEPLOYMENT_SUBSTAGE = "update_deployment"

	// RECONCILE_CREATE_SERVICE_SUBSTAGE is sub stage to create service.
	RECONCILE_CREATE_SERVICE_SUBSTAGE = "create_service"

	// RECONCILE_GET_SERVICE_SUBSTAGE is sub stage to get service.
	RECONCILE_GET_SERVICE_SUBSTAGE = "get_service"

	// RECONCILE_LIST_SERVICES_SUBSTAGE is sub stage to list service.
	RECONCILE_LIST_SERVICES_SUBSTAGE = "list_service"

	// RECONCILE_UPDATE_SERVICE_SUBSTAGE is sub stage to update service.
	RECONCILE_UPDATE_SERVICE_SUBSTAGE = "update_service"

	// RECONCILE_CREATE_OTHER_SERVICE_SUBSTAGE is sub stage to create other service.
	RECONCILE_CREATE_OTHER_SERVICE_SUBSTAGE = "create_other_service"

	// RECONCILE_UPDATE_OTHER_SERVICE_SUBSTAGE is sub stage to update other service.
	RECONCILE_UPDATE_OTHER_SERVICE_SUBSTAGE = "update_other_service"

	// RECONCILE_DELETE_OTHER_SERVICE_SUBSTAGE is sub stage to delete other service.
	RECONCILE_DELETE_OTHER_SERVICE_SUBSTAGE = "delete_other_service"

	// RECONCILE_LIST_PODS_SUBSTAGE is sub stage to list pods.
	RECONCILE_LIST_PODS_SUBSTAGE = "list_pods"

	// RECONCILE_UPDATE_BOOT_META_SUBSTAGE is sub stage to update boot metadata.
	RECONCILE_UPDATE_BOOT_META_SUBSTAGE = "update_boot_meta"
)

var (
	// ReconcileErrors is a prometheus counter metrics which holds the total
	// number of errors from the logan Reconciler
	ReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "logan_controller_runtime_reconcile_errors_total",
		Help: "Total number of logan reconciliation errors per controller",
	}, []string{"kind", "stage", "sub_stage", "boot"})

	// ReconcileTime is a prometheus metric which keeps track of the duration
	// of logan reconciliations
	ReconcileTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "logan_controller_runtime_reconcile_time_seconds",
		Help: "Length of time per logan reconciliation per controller",
	}, []string{"kind"})
)

func init() {
	metrics.Registry.MustRegister(
		ReconcileErrors,
		ReconcileTime,
	)
}

// UpdateReconcileTime will update reconcile time for each reconcile
func UpdateReconcileTime(kind string, reconcileTime time.Duration) {
	ReconcileTime.WithLabelValues(kind).Observe(reconcileTime.Seconds())
}

// UpdateReconcileErrors will update reconcile error metrics for sub/main stage
func UpdateReconcileErrors(kind string, stage string, subStage string, boot string) {
	ReconcileErrors.WithLabelValues(kind, stage, subStage, boot).Inc()
}

// UpdateMainStageErrors will update reconcile error metrics only for main stage
func UpdateMainStageErrors(kind string, stage string, boot string) {
	ReconcileErrors.WithLabelValues(kind, stage, "", boot).Inc()
}
