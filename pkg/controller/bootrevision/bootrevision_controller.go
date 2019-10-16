package bootrevision

import (
	"context"
	"github.com/go-logr/logr"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	loganMetrics "github.com/logancloud/logan-app-operator/pkg/logan/metrics"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
	"time"
)

var log = logf.Log.WithName("logan_controller_bootrevision")
var kindType = "BootRevision"

// Add creates a new BootRevision Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBootRevision{
		client:   util.NewClient(mgr.GetClient()),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("bootRevision-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("bootrevision-controller", mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: logan.MaxConcurrentReconciles})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BootRevision
	err = c.Watch(&source.Kind{Type: &appv1.BootRevision{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBootRevision implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBootRevision{}

// ReconcileBootRevision reconciles a BootRevision object
type ReconcileBootRevision struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   util.K8SClient
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a BootRevision object and makes changes based on the state read
// and what is in the BootRevision.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBootRevision) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("bootrevision", request)

	if operator.Ignore(request.Namespace) {
		return reconcile.Result{}, nil
	}

	logger.Info("Reconciling BootRevision")
	// Update metrics after processing each Reconcile
	reconcileStartTS := time.Now()
	defer func() {
		loganMetrics.UpdateReconcileTime(kindType, time.Now().Sub(reconcileStartTS))
	}()

	// Fetch the BootRevision instance
	instance := &appv1.BootRevision{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("BootRevision resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get BootRevision")
		return reconcile.Result{}, err
	}

	// all ok
	if instance.OwnerReferences != nil {
		return reconcile.Result{}, nil
	}

	found, boot := getBoot(instance, r.client, logger)
	if found {
		logger.Info("Found the boot.Set Controller Reference for revision",
			"instance", instance, "boot", boot)

		err := controllerutil.SetControllerReference(boot, instance, r.scheme)
		if err != nil {
			logger.Error(err, "Set Controller Reference for revision with error",
				"instance", instance, "boot", boot)
			return reconcile.Result{Requeue: true}, err
		}

		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			logger.Error(err, "Update revision with error",
				"instance", instance)
			return reconcile.Result{Requeue: true}, err
		}
	} else {
		logger.Info("Can not find boot for revision", "instance", instance)
		retryStr := instance.Annotations[keys.BootRevisionRetryAnnotationKey]
		retry, _ := strconv.Atoi(retryStr)

		if retry > 20 {
			logger.Info("Maximum number of retries for revision. Delete it!",
				"instance", instance)
			err := r.client.Delete(context.TODO(), instance)
			if err != nil {
				logger.Error(err, "Can not delete revision.",
					"bootRevision", instance)
				return reconcile.Result{Requeue: true}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}

		retry += 1
		instance.Annotations[keys.BootRevisionRetryAnnotationKey] = strconv.Itoa(retry)
		logger.Info("Update revision retry times.", "instance", instance)
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			logger.Error(err, "Update revision for retry with error",
				"instance", instance)
			return reconcile.Result{Requeue: true}, err
		}
	}

	return reconcile.Result{Requeue: true}, nil
}

func getBoot(revision *appv1.BootRevision, client client.Client, logger logr.Logger) (bool, metav1.Object) {
	if revision.Labels == nil {
		return false, nil
	}

	nn := types.NamespacedName{
		Namespace: revision.Namespace,
		Name:      revision.Labels[keys.BootNameKey],
	}

	bootType := revision.Labels[keys.BootTypeKey]
	if bootType == logan.BootJava {
		javaBoot := &appv1.JavaBoot{}
		err := client.Get(context.TODO(), nn, javaBoot)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Boot resource not found.Maybe it hasn't been created yet.")
			} else {
				logger.Error(err, "Failed to get Boot")
			}
			return false, nil
		}
		return true, javaBoot
	} else if bootType == logan.BootPhp {
		phpBoot := &appv1.PhpBoot{}
		err := client.Get(context.TODO(), nn, phpBoot)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Boot resource not found.Maybe it hasn't been created yet.")
			} else {
				logger.Error(err, "Failed to get Boot")
			}
			return false, nil
		}
		return true, phpBoot
	} else if bootType == logan.BootPython {
		pythonBoot := &appv1.PythonBoot{}
		err := client.Get(context.TODO(), nn, pythonBoot)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Boot resource not found.Maybe it hasn't been created yet.")
			} else {
				logger.Error(err, "Failed to get Boot")
			}
			return false, nil
		}
		return true, pythonBoot
	} else if bootType == logan.BootNodeJS {
		nodejsBoot := &appv1.NodeJSBoot{}
		err := client.Get(context.TODO(), nn, nodejsBoot)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Boot resource not found.Maybe it hasn't been created yet.")
			} else {
				logger.Error(err, "Failed to get Boot")
			}
			return false, nil
		}
		return true, nodejsBoot
	} else if bootType == logan.BootWeb {
		webBoot := &appv1.WebBoot{}
		err := client.Get(context.TODO(), nn, webBoot)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Boot resource not found.Maybe it hasn't been created yet.")
			} else {
				logger.Error(err, "Failed to get Boot")
			}
			return false, nil
		}
		return true, webBoot
	}

	return false, nil
}
