package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Testing Volume", func() {
	var bootKey types.NamespacedName
	var phpBoot *bootv1.PhpBoot

	BeforeEach(func() {
		// Gen new namespace
		bootKey = operatorFramework.GenResource()
		operatorFramework.CreateNamespace(bootKey.Namespace)

		phpBoot = operatorFramework.SamplePhpBoot(bootKey)
	})

	AfterEach(func() {
		// Clean namespace
		operatorFramework.DeleteNamespace(bootKey.Namespace)
	})

	It("Test persistentVolumeClaim decode ", func() {
		(&(operatorFramework.E2E{
			Build: func() {
				if phpBoot.ObjectMeta.Annotations == nil {
					phpBoot.ObjectMeta.Annotations = make(map[string]string)
				}
				phpBoot.ObjectMeta.Annotations[config.BootProfileAnnotationKey] = "vol"
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
})
