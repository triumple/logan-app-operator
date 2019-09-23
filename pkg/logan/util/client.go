package util

import (
	"context"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8SClient is a K8S's client wrapper
type K8SClient struct {
	client.Client
}

// NewClient return a K8S's client wrapper
func NewClient(c client.Client) K8SClient {
	return K8SClient{c}
}

// ListRevision get a revision list by LabelSelector from namespace
func (k8s *K8SClient) ListRevision(namespace string, ls map[string]string) (*v1.BootRevisionList, error) {
	revisionList := &v1.BootRevisionList{}
	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(ls),
	}

	err := k8s.List(context.TODO(), listOptions, revisionList)
	if err != nil {
		return nil, err
	}
	return revisionList, nil
}
