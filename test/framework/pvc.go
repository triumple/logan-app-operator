package framework

import (
	"context"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

// GetPvcName will return specific pvc name with boot name
func GetPvcName(bootKey types.NamespacedName, isShared bool) string {
	pvcType := "nas"
	if isShared {
		pvcType = "shared-nas"
	}

	return bootKey.Name + "-" + pvcType
}

// GetPvcNameWithName will return specific pvc name with specific boot name
func GetPvcNameWithName(name string, isShared bool) string {
	pvcType := "nas"
	if isShared {
		pvcType = "shared-nas"
	}

	return name + "-" + pvcType
}

// GetEnvPvcName will return specific pvc name with env ${APP}
func GetEnvPvcName(isShared bool) string {
	pvcType := "nas"
	if isShared {
		pvcType = "shared-nas"
	}

	return "${APP}-" + pvcType
}

// SamplePvc will return specific pvc object according to boot key and isShared
func SamplePvc(bootKey types.NamespacedName, isShared bool) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app":      "havok",
		"bootName": bootKey.Name,
		"bootType": "php",
	}
	pvcNameConverted := GetPvcName(bootKey, isShared)

	return samplePvc(bootKey, isShared, labels, pvcNameConverted)
}

// SamplePvcWithName will return specific pvc object according to pvc name
func SamplePvcWithName(bootKey types.NamespacedName, isShared bool, pvcName string) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app":      "havok",
		"bootName": bootKey.Name,
		"bootType": "php",
	}
	pvcNameConverted := GetPvcNameWithName(pvcName, isShared)

	return samplePvc(bootKey, isShared, labels, pvcNameConverted)
}

// SamplePvcWithLabels will return specific pvc object according to labels
func SamplePvcWithLabels(bootKey types.NamespacedName, isShared bool, labels map[string]string) *corev1.PersistentVolumeClaim {
	pvcName := GetPvcName(bootKey, isShared)
	return samplePvc(bootKey, isShared, labels, pvcName)
}

func samplePvc(bootKey types.NamespacedName, isShared bool, labels map[string]string, pvcName string) *corev1.PersistentVolumeClaim {
	accessMode := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: bootKey.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessMode,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewScaledQuantity(1, resource.Giga),
				},
			},
		},
	}
	return pvc
}

// CreatePvc will create specific pvc object
func CreatePvc(obj runtime.Object) {
	err := framework.Mgr.GetClient().Create(context.TODO(), obj)
	if apierrors.IsInvalid(err) {
		log.Printf("failed to create object, got an invalid object error: %s", err.Error())
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	WaitDefaultUpdate()
}

// DeletePvc will delete specific pvc object
func DeletePvc(obj runtime.Object) {
	gomega.Eventually(func() error {
		return framework.Mgr.GetClient().Delete(context.TODO(), obj)
	}, defaultTimeout).Should(gomega.Succeed())
}

// IsInBootPvc will judge whether pvc in boot PersistentVolumeClaimMount, and return the pvc
func IsInBootPvc(pvcName string, pvcs []bootv1.PersistentVolumeClaimMount) (bool, *bootv1.PersistentVolumeClaimMount) {
	for _, vol := range pvcs {
		if vol.Name == pvcName {
			return true, &vol
		}
	}
	return false, nil
}

// IsInDeploymentPvc will judge whether pvc in deployment Volume, and return the pvc
func IsInDeploymentPvc(pvcName string, pvcs []corev1.Volume) (bool, *corev1.Volume) {
	for _, vol := range pvcs {
		if vol.Name == pvcName {
			return true, &vol
		}
	}
	return false, nil
}
