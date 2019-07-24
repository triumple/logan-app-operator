package framework

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type OperatorClient struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*OperatorClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &OperatorClient{client}, nil
}

func setConfigDefaults(config *rest.Config) error {
	gv := bootv1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	var Scheme = runtime.NewScheme()
	var Codecs = serializer.NewCodecFactory(Scheme)
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
