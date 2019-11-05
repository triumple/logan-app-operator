package framework

import (
	"context"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

// SampleSecret will return specific secret object according to boot key
func SampleSecret(bootKey types.NamespacedName) *corev1.Secret {
	return SampleSecretWithName(bootKey, bootKey.Name)
}

// SampleSecretWithName will return specific secret object according to boot key and specific name
func SampleSecretWithName(bootKey types.NamespacedName, name string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Namespace: bootKey.Namespace,
			Name:      name,
			Annotations: map[string]string{
				keys.BootSecretAnnotaionKeyPrefix + name: "true",
			},
		},
		Data: map[string][]byte{"url": []byte("url"),
			"password": []byte("password"),
			"username": []byte("username"),
		},
	}
	return secret
}

// CreateSecret will create specific Secret object
func CreateSecret(obj runtime.Object) {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error: %s", err.Error())
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	WaitDefaultUpdate()
}
