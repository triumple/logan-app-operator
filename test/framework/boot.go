package framework

import (
	"context"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

// SampleBoot will return a sample JavaBoot object
func SampleBoot(bootKey types.NamespacedName) *bootv1.JavaBoot {
	replicas := int32(1)
	javaboot := &bootv1.JavaBoot{
		ObjectMeta: metav1.ObjectMeta{Name: bootKey.Name, Namespace: bootKey.Namespace},
		Spec: bootv1.BootSpec{
			Port:     8080,
			Replicas: &replicas,
			Image:    "logancloud/logan-javaboot-sample",
			Version:  "latest",
		},
	}
	return javaboot
}

// SamplePhpBoot will return a sample PhpBoot object
func SamplePhpBoot(bootKey types.NamespacedName) *bootv1.PhpBoot {
	replicas := int32(1)
	phpBoot := &bootv1.PhpBoot{
		ObjectMeta: metav1.ObjectMeta{Name: bootKey.Name, Namespace: bootKey.Namespace},
		Spec: bootv1.BootSpec{
			Replicas: &replicas,
			Image:    "logan-startkit-boot",
			Version:  "1.2.1",
		},
	}
	return phpBoot
}

// CreateBoot will create Boot in kubernetes
func CreateBoot(obj runtime.Object) {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error:")
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	WaitDefaultUpdate()
}

// CreateBootWithError will create Boot in kubernetes, return error if occur
func CreateBootWithError(obj runtime.Object) error {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error:")
		return err
	}
	return err
}

// UpdateBoot will update JavaBoot to kubernetes
func UpdateBoot(boot *bootv1.JavaBoot) {
	gomega.Eventually(func() error {
		latestBoot := GetBoot(types.NamespacedName{Name: boot.Name, Namespace: boot.Namespace})
		latestBoot.Spec = boot.Spec
		err := framework.Mgr.GetClient().Update(context.TODO(), latestBoot)
		if apierrors.IsConflict(err) {
			log.Printf("failed to update object, got an Conflict error: ")
		}
		if apierrors.IsInvalid(err) {
			log.Printf("failed to update object, got an invalid object error: ")
		}
		return err
	}, defaultTimeout, defaultWaitSec).Should(gomega.Succeed())
	WaitDefaultUpdate()
}

// UpdatePhpBoot will update PhpBoot to kubernetes
func UpdatePhpBoot(boot *bootv1.PhpBoot) {
	gomega.Eventually(func() error {
		latestBoot := GetPhpBoot(types.NamespacedName{Name: boot.Name, Namespace: boot.Namespace})
		latestBoot.ObjectMeta.Name = boot.Name
		latestBoot.ObjectMeta.Namespace = boot.Namespace
		latestBoot.Spec = boot.Spec
		err := framework.Mgr.GetClient().Update(context.TODO(), latestBoot)
		if apierrors.IsConflict(err) {
			log.Printf("failed to update object, got an Conflict error: ")
		}
		if apierrors.IsInvalid(err) {
			log.Printf("failed to update object, got an invalid object error: ")
		}
		return err
	}, defaultTimeout, defaultWaitSec).Should(gomega.Succeed())
	WaitDefaultUpdate()
}

// UpdateBootWithError will update JavaBoot to kubernetes, return error if occur
func UpdateBootWithError(boot *bootv1.JavaBoot) error {
	err := framework.Mgr.GetClient().Update(context.TODO(), boot)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to update object, got an invalid object error: ")
		return err
	}
	return err
}

// DeleteBoot will delete Boot from kubernetes
func DeleteBoot(obj runtime.Object) {
	gomega.Eventually(func() error {
		return framework.Mgr.GetClient().Delete(context.TODO(), obj)
	}, defaultTimeout).Should(gomega.Succeed())
}

// GetBoot will get JavaBoot with boot key from kubernetes, return JavaBoot
func GetBoot(bootKey types.NamespacedName) *bootv1.JavaBoot {
	boot := &bootv1.JavaBoot{}
	gomega.Eventually(func() error {
		return framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("javaboots").Do().Into(boot)
	}, defaultTimeout).
		Should(gomega.Succeed())
	return boot
}

// GetPhpBoot will get PhpBoot with boot key from kubernetes, return PhpBoot
func GetPhpBoot(bootKey types.NamespacedName) *bootv1.PhpBoot {
	boot := &bootv1.PhpBoot{}
	gomega.Eventually(func() error {
		return framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("phpboots").Do().Into(boot)
	}, defaultTimeout).
		Should(gomega.Succeed())
	return boot
}

// GetBootWithError will get JavaBoot with boot key from kubernetes, return JavaBoot and error
func GetBootWithError(bootKey types.NamespacedName) (*bootv1.JavaBoot, error) {
	boot := &bootv1.JavaBoot{}
	err := framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("javaboots").Do().Into(boot)
	return boot, err
}
