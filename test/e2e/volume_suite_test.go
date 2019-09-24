package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Testing Volume", func() {
	var bootKey types.NamespacedName
	var phpBoot *bootv1.PhpBoot
	var pvc *corev1.PersistentVolumeClaim

	BeforeEach(func() {
		// Gen new namespace
		bootKey = operatorFramework.GenResource()
		operatorFramework.CreateNamespace(bootKey.Namespace)

		phpBoot = operatorFramework.SamplePhpBoot(bootKey)
		if phpBoot.ObjectMeta.Annotations == nil {
			phpBoot.ObjectMeta.Annotations = make(map[string]string)
		}
		phpBoot.ObjectMeta.Annotations[config.BootProfileAnnotationKey] = "vol"

		pvc = operatorFramework.SamplePvc(bootKey, false)
		operatorFramework.CreatePvc(pvc)
	})

	AfterEach(func() {
		// Clean namespace
		operatorFramework.DeleteNamespace(bootKey.Namespace)
	})

	It("Test persistentVolumeClaim decode ", func() {
		(&(operatorFramework.E2E{
			Build: func() {
				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				deploy := operatorFramework.GetDeployment(bootKey)
				for _, vol := range deploy.Spec.Template.Spec.Volumes {
					if vol.Name == "private-data" {
						Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(bootKey.Name + "-nas"))
					}
				}
			},
		})).Run()
	})

	It("Test update no pvc boot with pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)
		(&(operatorFramework.E2E{
			Build: func() {
				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				// checkout deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(1))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				boot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// checkout deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))
			},
		})).Run()
	})

	It("Test update pvc boot with no pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)
		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				emptyPvc := make([]bootv1.PersistentVolumeClaimMount, 0)
				boot.Spec.Pvc = emptyPvc

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(1))
			},
		})).Run()
	})

	It("Test update pvc boot with another pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				boot.Spec.Pvc[0].Name = newPvcName
				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, vol := operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
		})).Run()
	})

	It("Test create pvc boot then add another pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				boot.Spec.Pvc = append(boot.Spec.Pvc, pvcObject)

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
		})).Run()
	})

	It("Test create two pvc boot then delete one pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				updatePvc := make([]bootv1.PersistentVolumeClaimMount, 0)
				for _, p := range boot.Spec.Pvc {
					if p.Name != newPvcName {
						updatePvc = append(updatePvc, p)
					}
				}
				boot.Spec.Pvc = updatePvc

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, _ = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
		})).Run()
	})

	It("Test update two pvc boot with another two pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		newBootKeyC := operatorFramework.GenResource()
		newPvcNameC := operatorFramework.GetPvcNameWithName(newBootKeyC.Name, false)
		newPvcC := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyC.Name)
		operatorFramework.CreatePvc(newPvcC)

		newBootKeyD := operatorFramework.GenResource()
		newPvcNameD := operatorFramework.GetPvcNameWithName(newBootKeyD.Name, false)
		newPvcD := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyD.Name)
		operatorFramework.CreatePvc(newPvcD)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				boot.Spec.Pvc[0].Name = newPvcNameC
				boot.Spec.Pvc[1].Name = newPvcNameD

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameC, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameD, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				//check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, vol := operatorFramework.IsInDeploymentPvc(newPvcNameC, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameC))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcNameD, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameD))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
		})).Run()
	})

	It("Test create two pvc boot then delete one and add two pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		newBootKeyC := operatorFramework.GenResource()
		newPvcNameC := operatorFramework.GetPvcNameWithName(newBootKeyC.Name, false)
		newPvcC := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyC.Name)
		operatorFramework.CreatePvc(newPvcC)

		newBootKeyD := operatorFramework.GenResource()
		newPvcNameD := operatorFramework.GetPvcNameWithName(newBootKeyD.Name, false)
		newPvcD := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyD.Name)
		operatorFramework.CreatePvc(newPvcD)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				updatePvc := make([]bootv1.PersistentVolumeClaimMount, 0)
				for _, p := range boot.Spec.Pvc {
					if p.Name != pvcName {
						updatePvc = append(updatePvc, p)
					}
				}
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      newPvcNameC,
					MountPath: "/var/logs3",
				}
				updatePvc = append(updatePvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcNameD,
					MountPath: "/var/logs4",
				}
				updatePvc = append(updatePvc, pvcObject)
				boot.Spec.Pvc = updatePvc

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameC, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameD, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, vol := operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcNameC, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameC))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcNameD, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameD))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(4))
			},
		})).Run()
	})

	It("Test create two pvc boot then add two pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, false)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, false)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, false, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		newBootKeyC := operatorFramework.GenResource()
		newPvcNameC := operatorFramework.GetPvcNameWithName(newBootKeyC.Name, false)
		newPvcC := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyC.Name)
		operatorFramework.CreatePvc(newPvcC)

		newBootKeyD := operatorFramework.GenResource()
		newPvcNameD := operatorFramework.GetPvcNameWithName(newBootKeyD.Name, false)
		newPvcD := operatorFramework.SamplePvcWithName(bootKey, false, newBootKeyD.Name)
		operatorFramework.CreatePvc(newPvcD)
		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      newPvcNameC,
					MountPath: "/var/logs3",
				}
				boot.Spec.Pvc = append(boot.Spec.Pvc, pvcObject)
				pvcObject = bootv1.PersistentVolumeClaimMount{
					Name:      newPvcNameD,
					MountPath: "/var/logs4",
				}
				boot.Spec.Pvc = append(boot.Spec.Pvc, pvcObject)

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameC, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcNameD, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)

				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcNameC, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameC))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcNameD, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcNameD))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(5))
			},
		})).Run()
	})

	It("Test update shared pvc boot with no pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, true)
		pvc = operatorFramework.SamplePvc(bootKey, true)
		operatorFramework.CreatePvc(pvc)
		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				emptyPvc := make([]bootv1.PersistentVolumeClaimMount, 0)
				boot.Spec.Pvc = emptyPvc

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(1))
			},
		})).Run()
	})

	It("Test update shared pvc boot with another pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, true)
		pvc = operatorFramework.SamplePvc(bootKey, true)
		operatorFramework.CreatePvc(pvc)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, true)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, true, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)

		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				boot.Spec.Pvc[0].Name = newPvcName
				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, _ = operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(false))

				hasPvc, vol := operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
		})).Run()
	})

	It("Test create shared pvc boot then add another pvc", func() {
		pvcName := operatorFramework.GetPvcName(bootKey, true)
		pvc = operatorFramework.SamplePvc(bootKey, true)
		operatorFramework.CreatePvc(pvc)

		newBootKey := operatorFramework.GenResource()
		newPvcName := operatorFramework.GetPvcNameWithName(newBootKey.Name, true)
		newPvc := operatorFramework.SamplePvcWithName(bootKey, true, newBootKey.Name)
		operatorFramework.CreatePvc(newPvc)
		(&(operatorFramework.E2E{
			Build: func() {
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      pvcName,
					MountPath: "/var/logs",
				}
				phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)

				operatorFramework.CreateBoot(phpBoot)
			},
			Check: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(2))
			},
			Update: func() {
				boot := operatorFramework.GetPhpBoot(bootKey)
				pvcObject := bootv1.PersistentVolumeClaimMount{
					Name:      newPvcName,
					MountPath: "/var/logs2",
				}
				boot.Spec.Pvc = append(boot.Spec.Pvc, pvcObject)

				operatorFramework.UpdatePhpBoot(boot)
			},
			Recheck: func() {
				// check boot
				boot := operatorFramework.GetPhpBoot(bootKey)
				hasPvc, _ := operatorFramework.IsInBootPvc(pvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				hasPvc, _ = operatorFramework.IsInBootPvc(newPvcName, boot.Spec.Pvc)
				Expect(hasPvc).Should(Equal(true))

				// check deployment
				deploy := operatorFramework.GetDeployment(bootKey)
				hasPvc, vol := operatorFramework.IsInDeploymentPvc(pvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(pvcName))

				hasPvc, vol = operatorFramework.IsInDeploymentPvc(newPvcName, deploy.Spec.Template.Spec.Volumes)
				Expect(hasPvc).Should(Equal(true))
				Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(newPvcName))

				// include shared-data
				Expect(len(deploy.Spec.Template.Spec.Volumes)).Should(Equal(3))
			},
		})).Run()
	})
})
