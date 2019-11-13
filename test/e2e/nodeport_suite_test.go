package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Testing nodePort Boot", func() {
	var bootKey types.NamespacedName
	var javaBoot *bootv1.JavaBoot

	BeforeEach(func() {
		// Gen new namespace, run in dev ENV
		bootKey = operatorFramework.GenResource()
		bootKey.Namespace = bootKey.Namespace + "-dev"
		operatorFramework.CreateNamespace(bootKey.Namespace)

		javaBoot = operatorFramework.SampleBoot(bootKey)
	})

	AfterEach(func() {
		// Clean namespace
		operatorFramework.DeleteNamespace(bootKey.Namespace)
	})

	Describe("testing update nodePort boot", func() {
		It("testing update nodePort from true to false[Slow]", func() {
			nodePortBootKey := types.NamespacedName{
				Name:      bootKey.Name + "-external",
				Namespace: bootKey.Namespace,
			}
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodePort = "true"
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.NodePort).Should(Equal("true"))

					service := operatorFramework.GetService(nodePortBootKey)
					Expect(service.Name).Should(Equal(nodePortBootKey.Name))
					Expect(service.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
					Expect(service.Spec.Ports[0].NodePort).Should(Equal(service.Spec.Ports[0].Port))
				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)

					boot.Spec.NodePort = "false"
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					_, err := operatorFramework.GetServiceWithError(nodePortBootKey)
					Expect(err).To(HaveOccurred())
				},
			})).Run()
		})

		It("testing update nodePort from false to true[Slow]", func() {
			nodePortBootKey := types.NamespacedName{
				Name:      bootKey.Name + "-external",
				Namespace: bootKey.Namespace,
			}
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodePort = "false"
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.NodePort).Should(Equal("false"))

					_, err := operatorFramework.GetServiceWithError(nodePortBootKey)
					Expect(err).To(HaveOccurred())
				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.NodePort = "true"
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					service := operatorFramework.GetService(nodePortBootKey)
					Expect(service.Name).Should(Equal(nodePortBootKey.Name))
					Expect(service.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
					Expect(service.Spec.Ports[0].NodePort).Should(Equal(service.Spec.Ports[0].Port))
				},
			})).Run()
		})
	})
})
