package javaboot

import (
	"context"
	"github.com/go-logr/logr"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
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
	"time"
	"fmt"
)

var log = logf.Log.WithName("logan_controller_javaboot")
var bootTimestamp map[string]int64
var bootAction map[string]string

// Add creates a new Boot Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileJavaBoot{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("javaboot-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("javaboot-controller", mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: logan.MaxConcurrentReconciles})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource JavaBoot
	err = c.Watch(&source.Kind{Type: &appv1.JavaBoot{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Modify this to be the types you create(Deployment and Service) that are owned by the primary resource
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1.JavaBoot{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1.JavaBoot{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileJavaBoot implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileJavaBoot{}

// ReconcileJavaBoot reconciles a JavaBoot object
type ReconcileJavaBoot struct {
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
func (r *ReconcileJavaBoot) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("javaboot", request)

	if operator.Ignore(request.Namespace) {
		return reconcile.Result{}, nil
	}

	logger.Info("Reconciling JavaBoot")

	var handler *operator.BootHandler

	// Fetch the Boot instance
	javaBoot := &appv1.JavaBoot{}
	err := r.client.Get(context.TODO(), request.NamespacedName, javaBoot)
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
	}
	if _, ok := bootTimestamp[javaBoot.Name]; !ok {
		if bootTimestamp == nil {
			bootTimestamp = make(map[string]int64, 0)
		}
		bootTimestamp[javaBoot.Name] = time.Now().UnixNano() / int64(time.Millisecond)
	}
	if _, ok := bootAction[javaBoot.Name]; !ok {
		if bootAction == nil {
			bootAction = make(map[string]string, 0)
		}
		bootAction[javaBoot.Name] = "begin,"
	}
	handler = InitHandler(javaBoot, r.scheme, r.client, logger, r.recorder)

	changed := handler.DefaultValue()

	//Update the Boot's default Value
	if changed {
		reason := "Updating Boot with Defaulters"
		bootAction[javaBoot.Name] = bootAction[javaBoot.Name] + "UpdateDefault,"
		logger.Info(reason)
		err = r.client.Update(context.TODO(), javaBoot)
		if err != nil {
			logger.Info("Failed to update Boot", "boot", javaBoot)
			handler.EventFail(reason, javaBoot.Name, err)
			return reconcile.Result{Requeue: true}, nil
		}
		handler.EventNormal(reason, javaBoot.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// 1. Check the existence of components, if not exist, create new one.
	result, requeue, err := handler.ReconcileCreate()
	if requeue {
		bootAction[javaBoot.Name] = bootAction[javaBoot.Name] + "ReconcileCreate,"
		handler.DeployStatis(javaBoot.Name, javaBoot.Namespace, handler.Client)
		//handler.ServiceStatis(javaBoot.Name, javaBoot.Namespace, handler.Client)
		return result, err
	}


	// 2. Handle the update logic of components
	result, requeue, err = handler.ReconcileUpdate()
	if requeue {
		bootAction[javaBoot.Name] = bootAction[javaBoot.Name] + "ReconcileUpdate,"
		return result, err
	}

	result, requeue, updated, err := handler.ReconcileUpdateBootMeta()

	if updated {
		reason := "Updating Boot Meta"
		bootAction[javaBoot.Name] = bootAction[javaBoot.Name] + "UpdateBootMeta,"
		logger.Info(reason, "new", javaBoot.Annotations)
		err := r.client.Update(context.TODO(), javaBoot)
		if err != nil {
			// Other place will modify the status? So it will sometimes occur.
			logger.Info("Failed to update Boot Metadata", "err", err.Error())

			handler.EventFail(reason, javaBoot.Name, err)
			return reconcile.Result{Requeue: true}, nil
		}
		handler.EventNormal(reason, javaBoot.Name)
	}
	if requeue {
		return result, err
	}
	if _, ok := bootTimestamp[javaBoot.Name]; ok {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		totalTime := endTime - bootTimestamp[javaBoot.Name]
		s := fmt.Sprintf("java boot reconcile finish-%d-action-%s", totalTime, bootAction[javaBoot.Name])
		logger.Info("statics", "finish", s)
		delete(bootTimestamp, javaBoot.Name)
		delete(bootAction, javaBoot.Name)
	}

	return reconcile.Result{}, nil
}

// InitHandler will create the Handler for handling logic of Boot
func InitHandler(javaBoot *appv1.JavaBoot, scheme *runtime.Scheme,
	client client.Client, logger logr.Logger, recorder record.EventRecorder) (handler *operator.BootHandler) {
	boot := javaBoot.DeepCopyBoot()

	bootCfg := config.JavaConfig
	profileConfig, err := operator.GetProfileBootConfig(boot, logger)
	if err != nil {
		logger.Info(err.Error())
	} else if profileConfig != nil {
		bootCfg = profileConfig
	}

	return &operator.BootHandler{
		OperatorBoot: javaBoot,
		OperatorSpec: &javaBoot.Spec,
		OperatorMeta: &javaBoot.ObjectMeta,

		Boot:     boot,
		Config:   bootCfg,
		Scheme:   scheme,
		Client:   client,
		Logger:   logger,
		Recorder: recorder,
	}
}
