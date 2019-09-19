package framework

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace will create specific namespace in kubernetes
func CreateNamespace(name string) (*v1.Namespace, error) {
	namespace, err := framework.KubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create namespace with name %v", name))
	}
	return namespace, nil
}

// DeleteNamespace will delete specific namespace from kubernetes
func DeleteNamespace(name string) {
	option := &metav1.DeleteOptions{}
	err := framework.KubeClient.CoreV1().Namespaces().Delete(name, option)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
