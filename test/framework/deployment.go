package framework

import (
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

func CreateDeployment(dep *appsv1.Deployment) *appsv1.Deployment {
	deploy := &appsv1.Deployment{}
	var err error
	Eventually(func() error {
		deploy, err = framework.KubeClient.AppsV1().Deployments(dep.Namespace).Create(dep)
		return err
	}, defaultTimeout).
		Should(Succeed())
	WaitDefaultUpdate()
	return deploy
}

func CreateDeploymentWithError(dep *appsv1.Deployment) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	var err error
	Eventually(func() error {
		deploy, err = framework.KubeClient.AppsV1().Deployments(dep.Namespace).Create(dep)
		return err
	}, defaultTimeout).
		ShouldNot(Succeed())
	WaitDefaultUpdate()
	return deploy, err
}

func GetDeployment(nn types.NamespacedName) *appsv1.Deployment {
	deploy := &appsv1.Deployment{}
	var err error
	Eventually(func() error {
		deploy, err = framework.KubeClient.AppsV1().Deployments(nn.Namespace).Get(nn.Name, metav1.GetOptions{})
		return err
	}, defaultTimeout).
		Should(Succeed())
	return deploy
}

func UpdateDeployment(dep *appsv1.Deployment) *appsv1.Deployment {
	deploy := &appsv1.Deployment{}
	var err error
	Eventually(func() error {
		latest := GetDeployment(types.NamespacedName{Namespace: dep.Namespace, Name: dep.Name})
		latest.Spec = dep.Spec
		deploy, err = framework.KubeClient.AppsV1().Deployments(dep.Namespace).Update(latest)
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
	return deploy
}

func DeleteDeployment(dep *appsv1.Deployment) {
	Eventually(func() error {
		return framework.KubeClient.AppsV1().Deployments(dep.Namespace).Delete(dep.Name, &metav1.DeleteOptions{})
	}, defaultTimeout).
		Should(Succeed())
}
