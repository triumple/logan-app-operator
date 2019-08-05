package framework

import (
	"context"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

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

func CreateBoot(obj runtime.Object) {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error:")
		return
	}
	Expect(err).NotTo(HaveOccurred())
	WaitDefaultUpdate()
}

func CreateBootWithError(obj runtime.Object) error {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error:")
		return err
	}
	return err
}

func UpdateBoot(boot *bootv1.JavaBoot) {
	Eventually(func() error {
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
	}, defaultTimeout, defaultWaitSec).Should(Succeed())
	WaitDefaultUpdate()
}

func UpdatePhpBoot(boot *bootv1.PhpBoot) {
	Eventually(func() error {
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
	}, defaultTimeout, defaultWaitSec).Should(Succeed())
	WaitDefaultUpdate()
}

func UpdateBootWithError(boot *bootv1.JavaBoot) error {
	err := framework.Mgr.GetClient().Update(context.TODO(), boot)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to update object, got an invalid object error: ")
		return err
	}
	return err
}

func DeleteBoot(obj runtime.Object) {
	Eventually(func() error {
		return framework.Mgr.GetClient().Delete(context.TODO(), obj)
	}, defaultTimeout).Should(Succeed())
}

func GetBoot(bootKey types.NamespacedName) *bootv1.JavaBoot {
	boot := &bootv1.JavaBoot{}
	Eventually(func() error {
		return framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("javaboots").Do().Into(boot)
	}, defaultTimeout).
		Should(Succeed())
	return boot
}

func GetPhpBoot(bootKey types.NamespacedName) *bootv1.PhpBoot {
	boot := &bootv1.PhpBoot{}
	Eventually(func() error {
		return framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("phpboots").Do().Into(boot)
	}, defaultTimeout).
		Should(Succeed())
	return boot
}

func GetBootWithError(bootKey types.NamespacedName) (*bootv1.JavaBoot, error) {
	boot := &bootv1.JavaBoot{}
	err := framework.OperatorClient.restClient.Get().Namespace(bootKey.Namespace).Name(bootKey.Name).Resource("javaboots").Do().Into(boot)
	return boot, err
}
