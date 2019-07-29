package framework

import (
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

func GetService(nn types.NamespacedName) *corev1.Service {
	service := &corev1.Service{}
	var err error
	Eventually(func() error {
		service, err = framework.KubeClient.CoreV1().Services(nn.Namespace).Get(nn.Name, metav1.GetOptions{})
		return err
	}, defaultTimeout).
		Should(Succeed())
	WaitDefaultUpdate()
	return service
}

func UpdateService(svr *corev1.Service) *corev1.Service {
	service := &corev1.Service{}
	var err error
	Eventually(func() error {
		latest := GetService(types.NamespacedName{Name: svr.Name, Namespace: svr.Namespace})
		latest.Spec = svr.Spec
		service, err = framework.KubeClient.CoreV1().Services(svr.Namespace).Update(latest)
		if apierrors.IsConflict(err) {
			log.Printf("failed to update object, got an Conflict error: ")
		}
		if apierrors.IsInvalid(err) {
			log.Printf("failed to update object, got an invalid object error: ")
		}
		return err
	}, defaultTimeout).
		Should(Succeed())
	WaitDefaultUpdate()
	return service
}

func DeleteService(svr *corev1.Service) {
	Eventually(func() error {
		return framework.KubeClient.AppsV1().Deployments(svr.Namespace).Delete(svr.Name, &metav1.DeleteOptions{})
	}, defaultTimeout).
		Should(Succeed())
}
