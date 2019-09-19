package framework

import (
	"github.com/logancloud/logan-app-operator/pkg/apis"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

var defaultTimeout = 1 * time.Minute

// Framework is the struct for e2e test framework
type Framework struct {
	Mgr            manager.Manager
	KubeClient     kubernetes.Interface
	HTTPClient     *http.Client
	MasterHost     string
	DefaultTimeout time.Duration
	OperatorClient *OperatorClient
}

var (
	framework *Framework
)

// InitFramework will return framework object
func InitFramework() (*Framework, error) {
	var err error
	framework, err = New()
	return framework, err
}

// New will new the item defined in Framework structure
func New() (*Framework, error) {
	kubeconfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, errors.Wrap(err, "creating http-client failed")
	}
	mgr, err := manager.New(kubeconfig, manager.Options{
		Namespace:      "",
		MapperProvider: restmapper.NewDynamicRESTMapper,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating new manager failed")
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrap(err, "creating add to scheme failed")
	}

	restClient, err := NewForConfig(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "creating Rest-client failed")
	}

	f := &Framework{
		Mgr:            mgr,
		MasterHost:     kubeconfig.Host,
		KubeClient:     cli,
		OperatorClient: restClient,
		HTTPClient:     httpc,
		DefaultTimeout: time.Minute,
	}

	return f, nil
}
