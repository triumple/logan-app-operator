package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

var _ = Describe("Testing CRD [CRD]", func() {
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

	Describe("testing create boot name", func() {
		It("testing create ok boot name which starts with lower case alpha", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "abc"
					javaBoot.Name = bootKey.Name
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
			})).Run()
		})

		It("testing create error boot name which starts with upper case alpha", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "Abc"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot name which starts with number", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "99boot"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot name which not starts with lower case alpha", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "-abc"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create ok boot name which consist of lower case alphanumeric characters, case 1/2", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "ac12ab"
					javaBoot.Name = bootKey.Name
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
			})).Run()
		})

		It("testing create ok boot name which consist of lower case alphanumeric characters, case 2/2", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "ab--c1-2ab"
					javaBoot.Name = bootKey.Name
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
			})).Run()
		})

		It("testing create error boot name which not consist of lower case alphanumeric characters - case 1/2", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "a@b#c-d_e"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot name which not consist of lower case alphanumeric characters - case 2/2", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "abBcD"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot name which not end with an alphanumeric character", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "abc-"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create ok boot name with length of 47 characters", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "abcdef-123456-abcdef-123456-abcdef-123456-abcde"
					javaBoot.Name = bootKey.Name
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
				},
			})).Run()
		})

		It("testing create error boot name with length of 48 characters", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = "abcdef-123456-abcdef-123456-abcdef-123456-abcdef"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot name with length of 0 characters", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					bootKey.Name = ""
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot env", func() {
		Context("test create boot env name", func() {
			It("testing create ok boot env name which consist of alphabetic characters, digits, '_', '-', or '.', case 1/3", func() {
				envVar := corev1.EnvVar{
					Name:  ".aA.t_-8",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create ok boot env name which consist of alphabetic characters, digits, '_', '-', or '.', case 2/3", func() {
				envVar := corev1.EnvVar{
					Name:  "AaA.t_-8",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create ok boot env name which consist of alphabetic characters, digits, '_', '-', or '.', case 3/3", func() {
				envVar := corev1.EnvVar{
					Name:  "_AaA.t_-8",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create error boot env name which starts with number", func() {
				envVar := corev1.EnvVar{
					Name:  "8aAt._-8",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						javaBoot.Name = bootKey.Name
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("testing create error boot env name which consist of invalid characters(#)", func() {
				envVar := corev1.EnvVar{
					Name:  ".a#A.t_-*8",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						javaBoot.Name = bootKey.Name
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("testing create ok boot env name with length of 1 characters", func() {
				envVar := corev1.EnvVar{
					Name:  "E",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create error boot env name with length of 0 characters", func() {
				envVar := corev1.EnvVar{
					Name:  "",
					Value: "logan-startkit-boot-custom-env-test",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})
		})

		Context("test create boot env value", func() {
			It("testing create ok boot env value with string type, case 1/3", func() {
				envVar := corev1.EnvVar{
					Name:  ".aA.t_-8",
					Value: "true",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create ok boot env value with string type, case 2/3", func() {
				envVar := corev1.EnvVar{
					Name:  ".aA.t_-8",
					Value: "666.23",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})

			It("testing create ok boot env value with string type, case 3/3", func() {
				envVar := corev1.EnvVar{
					Name:  ".aA.t_-8",
					Value: "",
				}
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						isCustomEnvExist := false
						for _, value := range boot.Spec.Env {
							if value.Name == envVar.Name {
								isCustomEnvExist = true
								Expect(value.Value).Should(Equal(envVar.Value))
							}
						}
						Expect(isCustomEnvExist).Should(Equal(true))
					},
				})).Run()
			})
		})
	})

	Describe("testing create boot replicas", func() {
		It("testing create ok boot replicas 0", func() {
			replicas := int32(0)
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Replicas = &replicas
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Replicas).Should(Equal(replicas))
				},
			})).Run()
		})

		It("testing create ok boot replicas 100", func() {
			replicas := int32(100)
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Replicas = &replicas
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Replicas).Should(Equal(replicas))
				},
			})).Run()
		})

		It("testing create error boot replicas -1", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					replicas := int32(-1)
					javaBoot.Spec.Replicas = &replicas
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot replicas 101", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					replicas := int32(101)
					javaBoot.Spec.Replicas = &replicas
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot prometheus", func() {
		It("testing create ok boot prometheus 'true'", func() {
			prometheus := "true"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Prometheus = prometheus
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Prometheus).Should(Equal(prometheus))
				},
			})).Run()
		})

		It("testing create ok boot prometheus 'false'", func() {
			prometheus := "false"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Prometheus = prometheus
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Prometheus).Should(Equal(prometheus))
				},
			})).Run()
		})

		It("testing create ok boot prometheus empty", func() {
			prometheus := ""
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Prometheus = prometheus
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Prometheus).Should(Equal("true"))
				},
			})).Run()
		})

		It("testing create error boot prometheus invalid", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Prometheus = "abc"
					javaBoot.Name = bootKey.Name
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot health", func() {
		It("testing create ok boot health with length of 1", func() {
			health := "/"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Health).Should(Equal(health))
				},
			})).Run()
		})

		It("testing create ok boot health with length of 0", func() {
			health := ""
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Health).Should(Equal(health))
				},
			})).Run()
		})

		It("testing create ok boot health with length of 2048", func() {
			health := "/thousands of and hu years ago, in North America's past, all of its megafauna—large mammals such as mammoths and giant bears—disappeared. One proposed explanation for this event is that when the first Americans migrated over from Asia, they hunted the megafauna to extinction.These people, known as the Clovis society after a site where their distinctive spear points were first found, would have been able to use this food source to expand their population and fill the continent rapidly.Yet many scientists argue against this \"Pleistocene overkill\" hypothesis. Modern humans have certainly been capable of such drastic effects on animals, but could ancient people with little more than stone spears similarly have caused the extinction of numerous species of animals?Thirty-five genera or groups of species (and many individual species) suffered extinction in North America around 11,000 B.C., soon after the appearance and expansion of Paleo-lndians throughout the Americas (27 genera disappeared com$Thousands of years ago, in North America's past, all of its megafauna—large mammals such as mammoths and giant bears—disappeared. One proposed explanation for this event is that when the first Americans migrated over from Asia, they hunted the megafauna to extinction.These people, known as the Clovis society after a site where their distinctive spear points were first found, would have been able to use this food source to expand their population and fill the continent rapidly.Yet many scientists argue against this \"Pleistocene overkill\" hypothesis. Modern humans have certainly been capable of such drastic effects on animals, but could ancient people with little more than stone spears similarly have caused the extinction of numerous species of animals?Thirty-five genera or groups of species (and many individual species) suffered extinction in North America around 11,000 B.C., soon after the appearance and expansion of Paleo-lndians throughout the Americas (27 genera disappeared com$throughout the Americas (27genera disappeared c$"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Health).Should(Equal(health))
				},
			})).Run()
		})

		It("testing create ok boot health with special characters", func() {
			health := "/-h!@#e$%^&|t-+=h*)\\"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(*boot.Spec.Health).Should(Equal(health))
				},
			})).Run()
		})

		It("testing create error boot health with length of 2049", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					health := "/thousands of and hun years ago, in North America's past, all of its megafauna—large mammals such as mammoths and giant bears—disappeared. One proposed explanation for this event is that when the first Americans migrated over from Asia, they hunted the megafauna to extinction.These people, known as the Clovis society after a site where their distinctive spear points were first found, would have been able to use this food source to expand their population and fill the continent rapidly.Yet many scientists argue against this \"Pleistocene overkill\" hypothesis. Modern humans have certainly been capable of such drastic effects on animals, but could ancient people with little more than stone spears similarly have caused the extinction of numerous species of animals?Thirty-five genera or groups of species (and many individual species) suffered extinction in North America around 11,000 B.C., soon after the appearance and expansion of Paleo-lndians throughout the Americas (27 genera disappeared com$Thousands of years ago, in North America's past, all of its megafauna—large mammals such as mammoths and giant bears—disappeared. One proposed explanation for this event is that when the first Americans migrated over from Asia, they hunted the megafauna to extinction.These people, known as the Clovis society after a site where their distinctive spear points were first found, would have been able to use this food source to expand their population and fill the continent rapidly.Yet many scientists argue against this \"Pleistocene overkill\" hypothesis. Modern humans have certainly been capable of such drastic effects on animals, but could ancient people with little more than stone spears similarly have caused the extinction of numerous species of animals?Thirty-five genera or groups of species (and many individual species) suffered extinction in North America around 11,000 B.C., soon after the appearance and expansion of Paleo-lndians throughout the Americas (27 genera disappeared com$throughout the Americas (27genera disappeared c$"
					javaBoot.Spec.Health = &health
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot port", func() {
		It("testing create ok boot port 7788", func() {
			port := int32(7788)
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Port = port
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Port).Should(Equal(port))
				},
			})).Run()
		})

		It("testing create ok boot port 65535", func() {
			port := int32(65535)
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Port = port
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Port).Should(Equal(port))
				},
			})).Run()
		})

		It("testing create ok boot port 0[Slow]", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Port = 0
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Port).Should(Equal(int32(8080)))
				},
			})).Run()
		})

		It("testing create error boot port -1", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Port = -1
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})

		It("testing create error boot port 65536", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Port = 65536
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot resource", func() {
		Context("test create boot resource name", func() {
			It("testing create ok boot resource name cpu", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))
					},
				})).Run()
			})

			It("testing create ok boot resources name memory", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
					},
				})).Run()
			})

			It("testing create ok boot resources name storage", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceStorage] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						// todo not interface to get storage info
					},
				})).Run()
			})

			It("testing create ok boot resources name ephemeral-storage", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceEphemeralStorage] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						// todo not interface to get storage info
					},
				})).Run()
			})

			It("testing create ok boot resources name invalid", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits["abc"] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
					},
				})).Run()
			})
		})

		Context("test create boot resource cpu value", func() {
			It("testing create error boot resource value cpu num -1", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(-1, resource.DecimalSI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("testing create ok boot resource value cpu num 0", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(0, resource.DecimalSI)
				defaultCpuNum := *resource.NewQuantity(2, resource.DecimalSI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Value()).Should(Equal(defaultCpuNum.Value()))
					},
				})).Run()
			})

			It("testing create ok boot resource value cpu num 500", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(500, resource.DecimalSI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))
					},
				})).Run()
			})

			It("testing create ok boot resource value cpu num 500m", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(500, resource.DecimalSI)
				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))

						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))
					},
				})).Run()
			})
		})

		It("testing create ok boot resources name memory value 500", func() {
			resources := &corev1.ResourceRequirements{
				Limits:   map[corev1.ResourceName]resource.Quantity{},
				Requests: map[corev1.ResourceName]resource.Quantity{},
			}
			resources.Limits[corev1.ResourceMemory] = *resource.NewQuantity(500, resource.DecimalSI)
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Resources = *resources
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))

					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				},
			})).Run()
		})
	})

	Describe("testing create boot SessionAffinity", func() {
		It("testing create ok boot SessionAffinity 'ClientIP'", func() {
			sessionAffinity := "ClientIP"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.SessionAffinity = sessionAffinity
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.SessionAffinity).Should(Equal(sessionAffinity))
				},
			})).Run()
		})

		It("testing create ok boot SessionAffinity 'None'", func() {
			sessionAffinity := "None"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.SessionAffinity = sessionAffinity
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.SessionAffinity).Should(Equal(sessionAffinity))
				},
			})).Run()
		})

		It("testing create ok boot SessionAffinity empty", func() {
			sessionAffinity := ""
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.SessionAffinity = sessionAffinity
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.SessionAffinity).Should(Equal(sessionAffinity))
				},
			})).Run()
		})

		It("testing create error boot SessionAffinity invalid", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.SessionAffinity = "abc"
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot nodePort", func() {
		It("testing create ok boot nodePort 'true'", func() {
			nodePort := "true"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodePort = nodePort
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.NodePort).Should(Equal(nodePort))
				},
			})).Run()
		})

		It("testing create ok boot nodePort 'false'", func() {
			nodePort := "false"
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodePort = nodePort
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					boot := operatorFramework.GetBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.NodePort).Should(Equal(nodePort))
				},
			})).Run()
		})

		It("testing create error boot nodePort invalid", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodePort = "abc"
				},
				Check: func() {
					err := operatorFramework.CreateBootWithError(javaBoot)
					Expect(err).Should(HaveOccurred())
				},
			})).Run()
		})
	})

	Describe("testing create boot pvc", func() {
		//var bootKey types.NamespacedName
		var phpBoot *bootv1.PhpBoot
		var pvc *corev1.PersistentVolumeClaim

		BeforeEach(func() {
			// Gen new namespace
			// bootKey = operatorFramework.GenResource()
			// operatorFramework.CreateNamespace(bootKey.Namespace)

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
			// operatorFramework.DeleteNamespace(bootKey.Namespace)
		})

		Context("test create boot pvc name", func() {
			It("testing create boot with pvc name ok", func() {
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
				})).Run()
			})

			It("testing create error boot pvc name empty", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						pvc := bootv1.PersistentVolumeClaimMount{
							Name:      "",
							MountPath: "abc",
						}
						phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvc)
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(phpBoot)
						Expect(err).Should(HaveOccurred())
						errStr := "spec.pvc.name in body should be at least 1 chars long"
						Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
					},
				})).Run()
			})

			It("testing create error boot pvc name too long", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						pvc := bootv1.PersistentVolumeClaimMount{
							Name:      "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
							MountPath: "abc",
						}
						phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvc)
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(phpBoot)
						Expect(err).Should(HaveOccurred())
						errStr := "spec.pvc.name in body should be at most 63 chars long"
						Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
					},
				})).Run()
			})
		})

		Context("test create boot pvc readonly", func() {
			It("testing create boot pvc readonly true", func() {
				pvcName := operatorFramework.GetPvcName(bootKey, false)
				(&(operatorFramework.E2E{
					Build: func() {
						pvcObject := bootv1.PersistentVolumeClaimMount{
							Name:      pvcName,
							ReadOnly:  true,
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
				})).Run()
			})

			It("testing create boot pvc readonly false", func() {
				pvcName := operatorFramework.GetPvcName(bootKey, false)
				(&(operatorFramework.E2E{
					Build: func() {
						pvcObject := bootv1.PersistentVolumeClaimMount{
							Name:      pvcName,
							ReadOnly:  false,
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
				})).Run()
			})
		})

		Context("test create boot pvc mountPath", func() {
			It("testing create error boot pvc mountPath empty", func() {
				pvcName := operatorFramework.GetPvcName(bootKey, false)
				(&(operatorFramework.E2E{
					Build: func() {
						pvc := bootv1.PersistentVolumeClaimMount{
							Name:      pvcName,
							MountPath: "",
						}
						phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvc)
					},
					Check: func() {
						err := operatorFramework.CreateBootWithError(phpBoot)
						Expect(err).Should(HaveOccurred())
						errStr := "spec.pvc.mountPath in body should be at least 1 chars long"
						Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
					},
				})).Run()
			})
		})
	})
})
