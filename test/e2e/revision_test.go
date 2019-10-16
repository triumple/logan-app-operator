package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

var _ = Describe("Testing Boot Revision [Revision]", func() {
	var bootKey types.NamespacedName
	var javaBoot *bootv1.JavaBoot
	var k8sClient util.K8SClient
	var e2eCase *operatorFramework.E2E
	BeforeEach(func() {
		// Gen new namespace
		bootKey = operatorFramework.GenResource()
		operatorFramework.CreateNamespace(bootKey.Namespace)

		javaBoot = operatorFramework.SampleBoot(bootKey)
		k8sClient = util.NewClient(framework.Mgr.GetClient())

		e2eCase = &operatorFramework.E2E{
			Build: func() {
				operatorFramework.CreateBoot(javaBoot)
			},
			Check: func() {
				boot := operatorFramework.GetBoot(bootKey)
				Expect(boot.Name).Should(Equal(bootKey.Name))
				podLabels := operator.PodLabels(boot.DeepCopyBoot())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(1))

				r := lst.Items[0]
				Expect(*r.Spec.Replicas).ShouldNot(Equal(*boot.Spec.Replicas))
				r.Spec.Replicas = boot.Spec.Replicas
				Expect(r.Spec).Should(Equal(boot.Spec))
				Expect(r.Name).Should(Equal(bootKey.Name + "-1"))
				Expect(r.GetRevisionId()).Should(Equal(1))
				Expect(len(r.GetOwnerReferences())).Should(Equal(1))
				Expect(r.Annotations[keys.BootRevisionDiffAnnotationKey]).Should(Equal(""))
				Expect(r.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseActive), Equal(operator.RevisionPhaseRunning)))
			},
		}
	})

	AfterEach(func() {
		// Clean namespace
		operatorFramework.DeleteNamespace(bootKey.Namespace)
	})
	Context("test create the boot with revision", func() {
		It("testing create boot with revision is ok", func() {
			e2eCase.Run()
		})
	})

	Context("test update the boot with revision, revision do not increase", func() {
		It("testing update boot replica with revision", func() {
			e2eCase.Update = func() {
				boot := operatorFramework.GetBoot(bootKey)
				newReplica := int32(3)
				boot.Spec.Replicas = &newReplica
				operatorFramework.UpdateBoot(boot)
			}
			e2eCase.Recheck = func() {
				boot := operatorFramework.GetBoot(bootKey)
				Expect(boot.Name).Should(Equal(bootKey.Name))
				Expect(*boot.Spec.Replicas).Should(Equal(int32(3)))
				podLabels := operator.PodLabels(boot.DeepCopyBoot())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(1))

				r := lst.Items[0]
				Expect(r.Name).Should(Equal(bootKey.Name + "-1"))
				Expect(r.GetRevisionId()).Should(Equal(1))
				Expect(len(r.GetOwnerReferences())).Should(Equal(1))
				Expect(r.Annotations[keys.BootRevisionDiffAnnotationKey]).Should(Equal(""))
				Expect(r.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseActive), Equal(operator.RevisionPhaseRunning)))
			}

			e2eCase.Run()
		})

		It("testing update boot env with revision", func() {
			bizEnv := corev1.EnvVar{
				Name:  "LAST_DEPLOY",
				Value: "123",
			}
			e2eCase.Update = func() {
				boot := operatorFramework.GetBoot(bootKey)
				boot.Spec.Env = append(boot.Spec.Env, bizEnv)
				operatorFramework.UpdateBoot(boot)
			}
			e2eCase.Recheck = func() {
				boot := operatorFramework.GetBoot(bootKey)
				Expect(boot.Name).Should(Equal(bootKey.Name))
				Expect(boot.Spec.Env[len(boot.Spec.Env)-1]).Should(Equal(bizEnv))

				podLabels := operator.PodLabels(boot.DeepCopyBoot())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(1))
				r := lst.Items[0]
				Expect(r.Name).Should(Equal(bootKey.Name + "-1"))
				Expect(r.GetRevisionId()).Should(Equal(1))
				Expect(len(r.GetOwnerReferences())).Should(Equal(1))
				Expect(r.Annotations[keys.BootRevisionDiffAnnotationKey]).Should(Equal(""))
				Expect(r.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseActive), Equal(operator.RevisionPhaseRunning)))
			}
			e2eCase.Run()
		})
	})

	Context("test update the boot with revision, revision will increase", func() {
		It("testing update boot Resource with revision is ok", func() {

			e2eCase.Update = func() {
				boot := operatorFramework.GetBoot(bootKey)

				boot.Spec.Port = 8090
				operatorFramework.UpdateBoot(boot)
			}

			e2eCase.Recheck = func() {
				boot := operatorFramework.GetBoot(bootKey)
				Expect(boot.Name).Should(Equal(bootKey.Name))
				Expect(boot.Spec.Port).Should(Equal(int32(8090)))
				Expect(boot.Annotations[keys.BootRevisionIdAnnotationKey]).Should(Equal("2"))
				podLabels := operator.PodLabels(boot.DeepCopyBoot())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(2))
				latest := lst.SelectLatestRevision()
				Expect(latest.Name).Should(Equal(bootKey.Name + "-2"))
				Expect(latest.GetRevisionId()).Should(Equal(2))
				Expect(len(latest.GetOwnerReferences())).Should(Equal(1))
				Expect(latest.Annotations[keys.BootRevisionDiffAnnotationKey]).ShouldNot(Equal(""))
				Expect(latest.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseActive), Equal(operator.RevisionPhaseRunning)))

				previous := lst.Items[0]
				Expect(previous.Name).Should(Equal(bootKey.Name + "-1"))
				Expect(previous.GetRevisionId()).Should(Equal(1))
				Expect(len(previous.GetOwnerReferences())).Should(Equal(1))
				Expect(previous.Annotations[keys.BootRevisionDiffAnnotationKey]).Should(Equal(""))
				Expect(previous.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseComplete), Equal(operator.RevisionPhaseCancel)))

			}

			e2eCase.Run()
		})

		It("testing update boot Resource with revision only 10 max history", func() {

			e2eCase.Update = func() {
				for i := 0; i < 11; i++ {
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.Image = boot.Spec.Image + strconv.Itoa(i)
					operatorFramework.UpdateBoot(boot)
				}
			}

			e2eCase.Recheck = func() {
				boot := operatorFramework.GetBoot(bootKey)
				Expect(boot.Name).Should(Equal(bootKey.Name))
				podLabels := operator.PodLabels(boot.DeepCopyBoot())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(10))
				latest := lst.SelectLatestRevision()
				Expect(boot.Annotations[keys.BootRevisionIdAnnotationKey]).Should(Equal(latest.Annotations[keys.BootRevisionIdAnnotationKey]))
				//Expect(latest.Name).Should(Equal(bootKey.Name + "-2"))
				//Expect(latest.GetRevisionId()).Should(Equal(2))
				Expect(len(latest.GetOwnerReferences())).Should(Equal(1))
				Expect(latest.Annotations[keys.BootRevisionDiffAnnotationKey]).ShouldNot(Equal(""))
				Expect(latest.Annotations[keys.BootRevisionPhaseAnnotationKey]).Should(Or(Equal(operator.RevisionPhaseActive), Equal(operator.RevisionPhaseRunning)))
			}
			e2eCase.Run()
		})
	})

	Context("test delete the boot with revision", func() {
		It("test delete the boot with revision, revision should also deleted", func() {

			var podLabels map[string]string
			e2eCase.Update = func() {
				boot := operatorFramework.GetBoot(bootKey)
				podLabels = operator.PodLabels(boot.DeepCopyBoot())
				boot.Spec.Image = boot.Spec.Image + "123"
				operatorFramework.UpdateBoot(boot)

				boot = operatorFramework.GetBoot(bootKey)
				operatorFramework.DeleteBoot(boot)
				operatorFramework.WaitUpdate(20)
			}

			e2eCase.Recheck = func() {
				boot, err := operatorFramework.GetBootWithError(bootKey)
				Expect(err).Should(HaveOccurred())
				lst, err := k8sClient.ListRevision(boot.Namespace, podLabels)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(lst.Items)).Should(Equal(0))
			}

			e2eCase.Run()
		})
	})
})
