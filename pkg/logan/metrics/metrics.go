package metrics

import (
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconcileTotal is a prometheus counter metrics which holds the total
	// number of reconciliations per controller. It has two labels. controller label refers
	// to the controller name and result label refers to the reconcile result i.e
	// success, error, requeue, requeue_after
	ReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "logan_controller_runtime_reconcile_total",
		Help: "Total number of logan reconciliations per controller",
	}, []string{"controller", "result"})

	// ReconcileErrors is a prometheus counter metrics which holds the total
	// number of errors from the Reconciler
	ReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "logan_controller_runtime_reconcile_errors_total",
		Help: "Total number of logan reconciliation errors per controller",
	}, []string{"controller", "stage"})

	// ReconcileTime is a prometheus metric which keeps track of the duration
	// of reconciliations
	ReconcileTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "logan_controller_runtime_reconcile_time_seconds",
		Help: "Length of time per logan reconciliation per controller",
	}, []string{"controller"})
)

func Init() {
	metrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileErrors,
		ReconcileTime,
	)
}

// updateMetrics updates prometheus metrics within the controller
func UpdateReconcileTimeMetrics(reconcileTime time.Duration, controller string) {
	ReconcileTime.WithLabelValues(controller).Observe(reconcileTime.Seconds())
}

func UpdateReconcileErrorsMetrics(stage string, controller string) {
	ReconcileErrors.WithLabelValues(controller, stage).Inc()
}