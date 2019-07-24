package e2e

import (
	"context"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Testing CRD", func() {
	var bootKey types.NamespacedName
	var javaBoot *bootv1.JavaBoot
	BeforeEach(func() {
		// Gen new namespace
		bootKey = operatorFramework.GenResource()
		operatorFramework.CreateNamespace(bootKey.Namespace)
		javaBoot = operatorFramework.SampleBoot(bootKey)
	})

	AfterEach(func() {
		// Clean namespace
		operatorFramework.DeleteNamespace(bootKey.Namespace)
	})

	It("testing create error boot name", func() {
		(&(operatorFramework.E2E{
			Build: func() {
				bootKey.Name = "99boot"
				javaBoot.Name = bootKey.Name
			},
			Check: func() {
				err := framework.Mgr.GetClient().Create(context.TODO(), javaBoot)
				Expect(err).Should(HaveOccurred())
			},
		})).Run()
	})
})
