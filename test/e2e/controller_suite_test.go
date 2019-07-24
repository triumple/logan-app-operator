package e2e

import (
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
)

var _ = Describe("Testing Boot", func() {
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

	Context("test create boot", func() {
		It("testing create boot", func() {
			operatorFramework.CreateBoot(javaBoot)

			boot := operatorFramework.GetBoot(bootKey)
			Expect(boot.Name).Should(Equal(bootKey.Name))

			deploy := operatorFramework.GetDeployment(bootKey)
			operatorFramework.DeleteDeployment(deploy)

			svr := operatorFramework.GetService(bootKey)
			operatorFramework.DeleteService(svr)
		})
	})

	Describe("testing update boot", func() {
		It("testing update replicas", func() {

			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Replicas).Should(Equal(javaBoot.Spec.Replicas))
				},
				UpdateAndCheck: func() {
					newReplicas := int32(3)
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.Replicas = &newReplicas
					operatorFramework.UpdateBoot(boot)

					updateDeploy := operatorFramework.GetDeployment(bootKey)
					Expect(updateDeploy.Spec.Replicas).Should(Equal(&newReplicas))
				},
			})).Run()
		})

		Context("test update image version", func() {
			It("testing update version", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						image := javaBoot.Spec.Image + ":" + javaBoot.Spec.Version
						Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
					},
					UpdateAndCheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						boot.Spec.Version = "V1.0.1"
						operatorFramework.UpdateBoot(boot)

						updateDeploy := operatorFramework.GetDeployment(bootKey)
						updateImages := boot.Spec.Image + ":" + boot.Spec.Version
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(updateImages))
					},
				})).Run()
			})

			It("testing update image", func() {
				(&(operatorFramework.E2E{
					Build: func() {
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						image := javaBoot.Spec.Image + ":" + javaBoot.Spec.Version
						Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
					},
					UpdateAndCheck: func() {
						// update boot
						boot := operatorFramework.GetBoot(bootKey)
						boot.Spec.Image = "logan-startkit-boot-new"
						operatorFramework.UpdateBoot(boot)

						//recheck
						updateDeploy := operatorFramework.GetDeployment(bootKey)
						updateImages := boot.Spec.Image + ":" + boot.Spec.Version
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(updateImages))
					},
				})).Run()
			})
		})

		It("testing update port", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					svc := operatorFramework.GetService(bootKey)
					Expect(svc.Spec.Ports[0].Port).Should(Equal(javaBoot.Spec.Port))
					Expect(svc.Annotations["prometheus.io/port"]).Should(Equal(strconv.Itoa(8080)))
					Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(javaBoot.Spec.Port))
				},
				UpdateAndCheck: func() {
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.Port = 8081

					operatorFramework.UpdateBoot(boot)

					updateDeploy := operatorFramework.GetDeployment(bootKey)
					updatesvc := operatorFramework.GetService(bootKey)
					Expect(updatesvc.Spec.Ports[0].Port).Should(Equal(boot.Spec.Port))
					Expect(updatesvc.Annotations["prometheus.io/port"]).Should(Equal("8081"))
					Expect(updateDeploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(boot.Spec.Port))
					Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).Should(Equal(boot.Spec.Port))
					Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal).Should(Equal(boot.Spec.Port))
				},
			})).Run()
		})

		Context("testing update resources", func() {
			It("testing scale up cpu and memory", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}
				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)
				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

					},
					UpdateAndCheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						updateResources := &corev1.ResourceRequirements{
							Limits:   map[corev1.ResourceName]resource.Quantity{},
							Requests: map[corev1.ResourceName]resource.Quantity{},
						}

						updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
						updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

						updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
						updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

						boot.Spec.Resources = *updateResources
						operatorFramework.UpdateBoot(boot)

						//check resource
						updateDeploy := operatorFramework.GetDeployment(bootKey)
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Requests.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Requests.Cpu()))
					},
				})).Run()

			})

			It("testing scale down cpu and memory", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

					},
					UpdateAndCheck: func() {
						boot := operatorFramework.GetBoot(bootKey)

						updateResources := &corev1.ResourceRequirements{
							Limits:   map[corev1.ResourceName]resource.Quantity{},
							Requests: map[corev1.ResourceName]resource.Quantity{},
						}

						updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
						updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

						updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
						updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

						boot.Spec.Resources = *updateResources
						operatorFramework.UpdateBoot(boot)

						//recheck
						updateDeploy := operatorFramework.GetDeployment(bootKey)
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Requests.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Requests.Cpu()))

					},
				})).Run()
			})

			It("testing  cpu and memory Limits lager than Requests", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

					},
					UpdateAndCheck: func() {
						boot := operatorFramework.GetBoot(bootKey)

						updateResources := &corev1.ResourceRequirements{
							Limits:   map[corev1.ResourceName]resource.Quantity{},
							Requests: map[corev1.ResourceName]resource.Quantity{},
						}

						updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
						updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

						updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*6, resource.BinarySI)
						updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*6, resource.DecimalSI)

						boot.Spec.Resources = *updateResources
						operatorFramework.UpdateBoot(boot)

						//recheck
						updateDeploy := operatorFramework.GetDeployment(bootKey)

						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

						// Requests should equal or less than Limits
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Limits.Memory()))
						Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

					},
				})).Run()
			})
		})

		It("testing update nodeSelector", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					nodeSelector := map[string]string{"env": "dev", "app": "myAPPName"}
					javaBoot.Spec.NodeSelector = nodeSelector

					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaBoot.Spec.NodeSelector))

				},
				UpdateAndCheck: func() {
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.NodeSelector = map[string]string{"env": "test", "app": "myAPPName2", "new": "new_label"}

					operatorFramework.UpdateBoot(boot)

					//recheck
					updateDeploy := operatorFramework.GetDeployment(bootKey)
					Expect(updateDeploy.Spec.Template.Spec.NodeSelector).Should(Equal(boot.Spec.NodeSelector))
				},
			})).Run()
		})

		It("testing update health", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					health := "/health"
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(*javaBoot.Spec.Health))
					Expect(deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(*javaBoot.Spec.Health))

				},
				UpdateAndCheck: func() {
					boot := operatorFramework.GetBoot(bootKey)

					health2 := "/health2"
					boot.Spec.Health = &health2

					operatorFramework.UpdateBoot(boot)

					//recheck
					updateDeploy := operatorFramework.GetDeployment(bootKey)

					Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(health2))
					Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(health2))

				},
			})).Run()
		})

		It("testing update prometheusScrape", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Prometheus = "true"
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					svr := operatorFramework.GetService(bootKey)
					Expect(len(svr.Annotations)).Should(Equal(4))

				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)
					boot.Spec.Prometheus = "false"
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					updatesvc := operatorFramework.GetService(bootKey)
					Expect(len(updatesvc.Annotations)).Should(Equal(0))
				},
			})).Run()
		})

		It("testing update env simple", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					env := []corev1.EnvVar{
						{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"},
						{Name: "myApp", Value: "${APP}"},
						{Name: "myEnv", Value: "${ENV}"},
					}
					javaBoot.Spec.Env = env
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					for _, i := range javaBoot.Spec.Env {
						if strings.EqualFold(i.Name, "myAPP") {
							i.Value = bootKey.Name
						}
						if strings.EqualFold(i.Name, "myEnv") {
							i.Value = "test"
						}

						for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
							if strings.EqualFold(i.Name, j.Name) {
								Expect(i).Should(Equal(j))
							}
						}
					}
				},
				UpdateAndCheck: func() {
					boot := operatorFramework.GetBoot(bootKey)
					for _, i := range boot.Spec.Env {
						if strings.EqualFold(i.Name, "key1") {
							i.Value = "new_value"
						}
					}
					boot.Spec.Env = append(boot.Spec.Env, corev1.EnvVar{Name: "key5", Value: "value1"})
					operatorFramework.UpdateBoot(boot)

					updateDeploy := operatorFramework.GetDeployment(bootKey)

					for _, i := range boot.Spec.Env {
						found := false
						for _, j := range updateDeploy.Spec.Template.Spec.Containers[0].Env {
							if strings.EqualFold(i.Name, j.Name) {
								Expect(i).Should(Equal(j))
								found = true
							}
						}

						Expect(found).Should(Equal(true))
					}
				},
			})).Run()
		})

		It("testing set ownerReferences", func() {
			newReplicas := int32(3)

			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					svc := operatorFramework.GetService(bootKey)

					Expect(deploy.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
					Expect(deploy.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
					Expect(*deploy.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
					Expect(svc.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
					Expect(svc.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
					Expect(*svc.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
				},
				Update: func() {
					boot := operatorFramework.GetBoot(bootKey)

					boot.Spec.Replicas = &newReplicas
					operatorFramework.UpdateBoot(boot)
				},
				Recheck: func() {
					updateDeploy := operatorFramework.GetDeployment(bootKey)
					updateSvc := operatorFramework.GetService(bootKey)
					Expect(updateDeploy.Spec.Replicas).Should(Equal(&newReplicas))
					Expect(updateDeploy.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
					Expect(updateDeploy.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
					Expect(*updateDeploy.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
					Expect(updateSvc.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
					Expect(updateSvc.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
					Expect(*updateSvc.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
				},
			})).Run()
		})
	})

	Describe("can not update deployment by deployment", func() {
		It("can not update deployment replicas", func() {
			replicas := int32(1)
			e2e := &operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.Replicas = &replicas
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Replicas).Should(Equal(&replicas))
				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					newReplicas := int32(2)
					deploy.Spec.Replicas = &newReplicas
					operatorFramework.UpdateDeployment(deploy)
				},
			}
			e2e.Recheck = e2e.Check
			e2e.Run()
		})

		It("can not update deployment version", func() {
			e2e := &operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					image := javaBoot.Spec.Image + ":" + javaBoot.Spec.Version
					Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					deploy.Spec.Template.Spec.Containers[0].Image = "myImages"
					operatorFramework.UpdateDeployment(deploy)
				},
			}
			e2e.Recheck = e2e.Check

			e2e.Run()
		})

		It("can not update deployment port", func() {
			e2e := &(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					svr := operatorFramework.GetService(bootKey)
					Expect(svr.Spec.Ports[0].Port).Should(Equal(javaBoot.Spec.Port))
					Expect(svr.Annotations["prometheus.io/port"]).Should(Equal(strconv.Itoa(8080)))
					Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(javaBoot.Spec.Port))
					Expect(deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).Should(Equal(javaBoot.Spec.Port))
					Expect(deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal).Should(Equal(javaBoot.Spec.Port))
				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 8081
					svr := operatorFramework.GetService(bootKey)
					svr.Spec.Ports[0].Port = 8081
					svr.Annotations["prometheus.io/port"] = "8081"
					operatorFramework.UpdateDeployment(deploy)
					operatorFramework.UpdateService(svr)
				},
			})
			e2e.Recheck = e2e.Check
			e2e.Run()
		})

		Context("can not update resources by deployment", func() {
			It("can not scale up cpu and memory by deployment", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				e2e := &(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))
					},
					Update: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

						operatorFramework.UpdateDeployment(deploy)
					},
				})
				e2e.Recheck = e2e.Check

				e2e.Run()
			})

			It("can not scale down cpu and memory by deployment", func() {
				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				e2e := &(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Resources = *resources
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
						Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))
					},
					Update: func() {
						deploy := operatorFramework.GetDeployment(bootKey)
						deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
						deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

						operatorFramework.UpdateDeployment(deploy)
					},
				})
				e2e.Recheck = e2e.Check

				e2e.Run()
			})
		})

		It("can not update nodeSelector", func() {
			e2e := &operatorFramework.E2E{
				Build: func() {
					javaBoot.Spec.NodeSelector = map[string]string{"env": "test", "app": "myAPPName"}
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaBoot.Spec.NodeSelector))
				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					deploy.Spec.Template.Spec.NodeSelector = map[string]string{"env": "test", "app": "myAPPName2", "new": "new_label"}
					operatorFramework.UpdateDeployment(deploy)
				},
			}
			e2e.Recheck = e2e.Check
			e2e.Run()
		})

		It("can not update health", func() {
			e2e := &operatorFramework.E2E{
				Build: func() {
					health := "/health"
					javaBoot.Spec.Health = &health
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					Expect(deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(*javaBoot.Spec.Health))
					Expect(deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(*javaBoot.Spec.Health))

				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path = "/health2"
					deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path = "/health2"
					operatorFramework.UpdateDeployment(deploy)
				},
			}
			e2e.Recheck = e2e.Check
			e2e.Run()
		})

		It("can not update env simple", func() {
			e2e := &operatorFramework.E2E{
				Build: func() {

					javaBoot.Spec.Env = []corev1.EnvVar{
						{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"},
						{Name: "myApp", Value: "${APP}"},
						{Name: "myEnv", Value: "${ENV}"},
					}
					operatorFramework.CreateBoot(javaBoot)
				},
				Check: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					for _, i := range javaBoot.Spec.Env {
						if strings.EqualFold(i.Name, "myAPP") {
							i.Value = bootKey.Name
						}
						if strings.EqualFold(i.Name, "myEnv") {
							i.Value = "test"
						}

						for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
							if strings.EqualFold(i.Name, j.Name) {
								Expect(i).Should(Equal(j))
							}
						}
					}
				},
				Update: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					deploy.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
						{Name: "key1", Value: "value2"},
						{Name: "key5", Value: "value1"},
						{Name: "myApp", Value: "${APP}"},
						{Name: "myEnv", Value: "${ENV}"},
					}
					operatorFramework.UpdateDeployment(deploy)
				},
			}
			e2e.Recheck = e2e.Check
			e2e.Run()
		})
	})
})
