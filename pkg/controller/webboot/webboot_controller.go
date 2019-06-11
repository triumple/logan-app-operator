package webboot

import (
	"context"
	"github.com/go-logr/logr"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("logan_controller_webboot")

// Add creates a new Boot Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWebBoot{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("webboot-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("webboot-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource WebBoot
	err = c.Watch(&source.Kind{Type: &appv1.WebBoot{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Modify this to be the types you create(Deployment and Service) that are owned by the primary resource
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1.WebBoot{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1.WebBoot{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWebBoot implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWebBoot{}

// ReconcileWebBoot reconciles a WebBoot object
type ReconcileWebBoot struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Boot object and makes changes based on the state read
// and what is in the Boot.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWebBoot) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("webboot", request)

	if operator.Ignore(request.Namespace) {
		return reconcile.Result{}, nil
	}

	logger.Info("Reconciling WebBoot")

	var handler *operator.BootHandler

	// Fetch the Boot instance
	webBoot := &appv1.WebBoot{}
	err := r.client.Get(context.TODO(), request.NamespacedName, webBoot)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			logger.Info("Boot resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Boot")
		return reconcile.Result{}, err
	} else {
		handler = InitHandler(webBoot, r.scheme, r.client, logger, r.recorder)

		changed := handler.DefaultValue()

		//Update the Boot's default Value
		if changed {
			reason := "Updating Boot with Defaulters"
			logger.Info(reason)
			err = r.client.Update(context.TODO(), webBoot)
			if err != nil {
				logger.Info("Failed to update Boot", "boot", webBoot)
				handler.EventFail(reason, webBoot.Name, err)
				return reconcile.Result{Requeue: true}, nil
			}
			handler.EventNormal(reason, webBoot.Name)
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// 1. Check the existence of components, if not exist, create new one.
	result, err, requeue := handler.ReconcileCreate()
	if requeue {
		return result, err
	}

	// 2. Handle the update logic of components
	result, err, requeue = handler.ReconcileUpdate()
	if requeue {
		return result, err
	}

	result, err, requeue, updated := handler.ReconcileUpdateBootMeta()

	if updated {
		reason := "Updating Boot Meta"
		logger.Info(reason, "new", webBoot.Annotations)
		err := r.client.Update(context.TODO(), webBoot)
		if err != nil {
			// Other place will modify the status? So it will sometimes occur.
			logger.Info("Failed to update Boot Metadata", "err", err.Error())

			handler.EventFail(reason, webBoot.Name, err)
			return reconcile.Result{Requeue: true}, nil
		}
		handler.EventNormal(reason, webBoot.Name)
	}
	if requeue {
		return result, err
	}

	return reconcile.Result{}, nil
}

func InitHandler(webBoot *appv1.WebBoot, scheme *runtime.Scheme,
	client client.Client, logger logr.Logger, recorder record.EventRecorder) (handler *operator.BootHandler) {
	boot := webBoot.DeepCopyBoot()

	bootCfg := config.WebConfig
	profileConfig, err := operator.GetProfileBootConfig(boot, logger)
	if err != nil {
		logger.Info(err.Error())
	} else if profileConfig != nil {
		bootCfg = profileConfig
	}

	return &operator.BootHandler{
		OperatorBoot: webBoot,
		OperatorSpec: &webBoot.Spec,
		OperatorMeta: &webBoot.ObjectMeta,

		Boot:     boot,
		Config:   bootCfg,
		Scheme:   scheme,
		Client:   client,
		Logger:   logger,
		Recorder: recorder,
	}
}
