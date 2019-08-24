package e2e

import (
	ghodssyaml "github.com/ghodss/yaml"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Testing Webhook", func() {
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

	Context("test create the same boot name", func() {
		It("testing create same boot with same namespace and same kind", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
				Update: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create same boot with same namespace and different kind", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
				Update: func() {
					replicas := int32(1)
					phpboot := &bootv1.PhpBoot{
						ObjectMeta: metav1.ObjectMeta{Name: bootKey.Name, Namespace: bootKey.Namespace},
						Spec: bootv1.BootSpec{
							Replicas: &replicas,
							Image:    "logan-startkit-boot",
							Version:  "1.2.1",
						},
					}

					err := operatorFramework.CreateBootWithError(phpboot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create same boot with different namespace and same kind", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
				Update: func() {
					newBootKey := operatorFramework.GenResource()
					Expect(newBootKey.Namespace != javaBoot.Namespace).Should(BeTrue())

					javaBoot = operatorFramework.SampleBoot(newBootKey)
					javaBoot.Namespace = newBootKey.Namespace
					operatorFramework.CreateNamespace(newBootKey.Namespace)
					operatorFramework.CreateBoot(javaBoot)
					boot := operatorFramework.GetBoot(newBootKey)
					Expect(boot.Name).Should(Equal(newBootKey.Name))
				},
			})).Run()
		})

		It("testing create same boot with different namespace and different kind", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
				Update: func() {
					newBootKey := operatorFramework.GenResource()
					Expect(newBootKey.Namespace != bootKey.Namespace).Should(BeTrue())
					newBootKey.Name = bootKey.Name

					replicas := int32(1)
					phpboot := &bootv1.PhpBoot{
						ObjectMeta: metav1.ObjectMeta{Name: newBootKey.Name, Namespace: newBootKey.Namespace},
						Spec: bootv1.BootSpec{
							Replicas: &replicas,
							Image:    "logan-startkit-boot",
							Version:  "1.2.1",
						},
					}
					operatorFramework.CreateNamespace(newBootKey.Namespace)
					operatorFramework.CreateBoot(phpboot)
					boot := operatorFramework.GetPhpBoot(newBootKey)
					Expect(boot.Name).Should(Equal(newBootKey.Name))
				},
			})).Run()
		})
	})

	Describe("testing validating webhook", func() {
		It("check env with create operation", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					env := []corev1.EnvVar{
						{Name: "SPRING_ZIPKIN_ENABLED", Value: "false"},
					}

					javaBoot.Spec.Env = env
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
				Check: func() {
					_, err := operatorFramework.GetBootWithError(bootKey)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("check env with update operation", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					env := []corev1.EnvVar{
						{Name: "A", Value: "A"},
						{Name: "B", Value: "B"},
					}
					javaBoot.Spec.Env = env
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					_, err := operatorFramework.GetBootWithError(bootKey)
					Expect(err).Should(Succeed())
				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)
					for i, env := range boot.Spec.Env {
						if env.Name == "A" {
							boot.Spec.Env[i].Value = "new_A"
						}
					}
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					boot := operatorFramework.GetBoot(bootKey)
					found := false
					for _, env := range boot.Spec.Env {
						if env.Name == "A" {
							Expect(env.Value).Should(Equal("new_A"))
							found = true
						}
					}

					Expect(found).Should(Equal(true))
				},
			})).Run()
		})

		It("check env with update operation and empty env annotations", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					env := []corev1.EnvVar{
						{Name: "A", Value: "A"},
						{Name: "B", Value: "B"},
					}
					javaBoot.Spec.Env = env
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					_, err := operatorFramework.GetBootWithError(bootKey)
					Expect(err).Should(Succeed())
				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)
					for i, env := range boot.Spec.Env {
						if env.Name == "A" {
							boot.Spec.Env[i].Value = "new_A"
						}
					}
					delete(boot.Annotations, "app.logancloud.com/boot-envs")
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					boot := operatorFramework.GetBoot(bootKey)
					found := false
					for _, env := range boot.Spec.Env {
						if env.Name == "A" {
							Expect(env.Value).Should(Equal("new_A"))
							found = true
						}
					}

					Expect(found).Should(Equal(true))
				},
			})).Run()
		})

		Context("testing env with update operation annotation [Serial]", func() {
			var configNN = types.NamespacedName{
				Name:      "logan-app-operator-config",
				Namespace: "logan",
			}
			var env = corev1.EnvVar{
				Name:  "testKey",
				Value: "testValue",
			}
			BeforeEach(func() {
				// update config map
				c := operatorFramework.GetConfig(configNN)
				operator := c[logan.BootJava]

				operator.AppSpec.Env = append(operator.AppSpec.Env, env)
				c[logan.BootJava] = operator
				updatedConfigContent, _ := ghodssyaml.Marshal(&c)

				var configMap corev1.ConfigMap
				configMap.Namespace = configNN.Namespace
				configMap.Name = configNN.Name
				configMap.Data = make(map[string]string)
				configMap.Data["config.yaml"] = string(updatedConfigContent)
				operatorFramework.UpdateConfigmap(&configMap)
			})

			AfterEach(func() {
				// Clean config map
				c := operatorFramework.GetConfig(configNN)
				operator := c[logan.BootJava]

				tmp := operator.AppSpec.Env[:0]
				for _, value := range operator.AppSpec.Env {
					if value.Name != env.Name {
						tmp = append(tmp, value)
					}
				}
				operator.AppSpec.Env = tmp
				c[logan.BootJava] = operator
				updatedConfigContent, _ := ghodssyaml.Marshal(&c)

				var configMap corev1.ConfigMap
				configMap.Namespace = configNN.Namespace
				configMap.Name = configNN.Name
				configMap.Data = make(map[string]string)
				configMap.Data["config.yaml"] = string(updatedConfigContent)
				operatorFramework.UpdateConfigmap(&configMap)
			})

			It("check modify/add env with update operation annotation", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot, err := operatorFramework.GetBootWithError(bootKey)
						Expect(err).Should(Succeed())
						Expect(boot.Name).Should(Equal(bootKey.Name))
					},
					Update: func() {
						boot := operatorFramework.GetBoot(bootKey)
						for i, bootEnv := range boot.Spec.Env {
							if bootEnv.Name == env.Name {
								boot.Spec.Env[i].Value = "new_A"
							}
						}
						err := operatorFramework.UpdateBootWithError(boot)
						Expect(err).Should(HaveOccurred())
					},
					Recheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						found := false
						for _, bootEnv := range boot.Spec.Env {
							if bootEnv.Name == env.Name {
								Expect(env.Value).Should(Equal(env.Value))
								found = true
							}
						}

						Expect(found).Should(Equal(true))
					},
				})).Run()
			})

			It("check delete env with update operation annotation", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot, err := operatorFramework.GetBootWithError(bootKey)
						Expect(err).Should(Succeed())
						Expect(boot.Name).Should(Equal(bootKey.Name))
					},
					Update: func() {
						boot := operatorFramework.GetBoot(bootKey)

						tmp := boot.Spec.Env[:0]
						for _, value := range boot.Spec.Env {
							if value.Name != env.Name {
								tmp = append(tmp, value)
							}
						}
						boot.Spec.Env = tmp
						err := operatorFramework.UpdateBootWithError(boot)
						Expect(err).Should(HaveOccurred())
					},
					Recheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						found := false
						for _, bootEnv := range boot.Spec.Env {
							if bootEnv.Name == env.Name {
								Expect(env.Value).Should(Equal(env.Value))
								found = true
							}
						}

						Expect(found).Should(Equal(true))
					},
				})).Run()
			})
		})
	})

	Context("testing config reload", func() {
		configNN := types.NamespacedName{
			Name:      "logan-app-operator-config",
			Namespace: "logan"}

		It("config.yaml not found", func() {
			(&(operatorFramework.E2E{
				Update: func() {
					configmap := operatorFramework.GetConfigmap(configNN)
					delete(configmap.Data, "config.yaml")
					_, err := framework.KubeClient.CoreV1().ConfigMaps(configmap.Namespace).Update(configmap)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("config.yaml is blank", func() {
			(&(operatorFramework.E2E{
				Update: func() {
					configmap := operatorFramework.GetConfigmap(configNN)
					configmap.Data["config.yaml"] = ""
					_, err := framework.KubeClient.CoreV1().ConfigMaps(configmap.Namespace).Update(configmap)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
		It("config.yaml error format", func() {
			(&(operatorFramework.E2E{
				Update: func() {
					configmap := operatorFramework.GetConfigmap(configNN)
					configmap.Data["config.yaml"] = "{xx:123,}"
					_, err := framework.KubeClient.CoreV1().ConfigMaps(configmap.Namespace).Update(configmap)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})
})
