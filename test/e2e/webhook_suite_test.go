package e2e

import (
	"fmt"
	ghodssyaml "github.com/ghodss/yaml"
	bootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strings"
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

		Context("testing pvc with create operation", func() {
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
			})

			AfterEach(func() {
				// Clean namespace
				operatorFramework.DeleteNamespace(bootKey.Namespace)
			})
			Context("test create boot pvc webhook", func() {
				BeforeEach(func() {
					pvc = operatorFramework.SamplePvc(bootKey, false)
					operatorFramework.CreatePvc(pvc)
				})

				It("check nas pvc name with APP env ok", func() {
					pvcName := operatorFramework.GetEnvPvcName(false)
					expectPvcName := operatorFramework.GetPvcName(bootKey, false)
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
							hasPvc, vol := operatorFramework.IsInDeploymentPvc(expectPvcName, deploy.Spec.Template.Spec.Volumes)
							Expect(hasPvc).Should(Equal(true))
							Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(expectPvcName))
						},
					})).Run()
				})

				It("check nas pvc name with APP env too long", func() {
					addStr := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyza-nas"

					// 63 chars
					pvcName := "${APP}" + addStr
					expectPvcName := bootKey.Name + addStr

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc name %s must be not empty and no more than 63 characters", expectPvcName)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})

				It("check nas pvc name regex not match", func() {
					addStr := ".nas"
					pvcName := "${APP}" + addStr
					expectPvcName := bootKey.Name + addStr

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc %s is a DNS-1123 label.", expectPvcName)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})

				It("check shared pvc name with APP env ok", func() {
					pvc = operatorFramework.SamplePvc(bootKey, true)
					operatorFramework.CreatePvc(pvc)

					pvcName := operatorFramework.GetEnvPvcName(true)
					expectPvcName := operatorFramework.GetPvcName(bootKey, true)
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
							hasPvc, vol := operatorFramework.IsInDeploymentPvc(expectPvcName, deploy.Spec.Template.Spec.Volumes)
							Expect(hasPvc).Should(Equal(true))
							Expect(vol.PersistentVolumeClaim.ClaimName).Should(Equal(expectPvcName))
						},
					})).Run()
				})

				It("check shared pvc name with APP env too long", func() {
					addStr := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrst-shared-nas"

					// 63 chars
					pvcName := "${APP}" + addStr
					expectPvcName := bootKey.Name + addStr

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc name %s must be not empty and no more than 63 characters", expectPvcName)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})

				It("check shared pvc name regex not match", func() {
					addStr := ".shared.nas"
					pvcName := "${APP}" + addStr
					expectPvcName := bootKey.Name + addStr

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc %s is a DNS-1123 label.", expectPvcName)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})

				It("check pvc mountPath with ':'", func() {
					pvcName := operatorFramework.GetEnvPvcName(false)
					//expectPvcName := operatorFramework.GetPvcName(bootKey, false)
					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "a:b",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc MountPath must be not empty and  not contain ':'")
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})
			})

			Context("test create boot pvc label mismatch", func() {
				It("check nas pvc label mismatch case 1", func() {
					labels := map[string]string{
						"app":      "havok",
						"bootName": bootKey.Name,
					}
					pvc = operatorFramework.SamplePvcWithLabels(bootKey, false, labels)
					operatorFramework.CreatePvc(pvc)

					pvcName := operatorFramework.GetEnvPvcName(false)
					expectPvcName := operatorFramework.GetPvcName(bootKey, false)

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc %s's label don't match the boot %s", expectPvcName, bootKey.Name)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})

				It("check nas pvc label mismatch case 2", func() {
					labels := map[string]string{
						"app":      "havok",
						"bootName": bootKey.Name,
						"bootType": "php",
						"abc":      "123",
					}
					pvc = operatorFramework.SamplePvcWithLabels(bootKey, false, labels)
					operatorFramework.CreatePvc(pvc)

					pvcName := operatorFramework.GetEnvPvcName(false)
					expectPvcName := operatorFramework.GetPvcName(bootKey, false)

					(&(operatorFramework.E2E{
						Build: func() {
							pvcObject := bootv1.PersistentVolumeClaimMount{
								Name:      pvcName,
								MountPath: "/var/logs",
							}
							phpBoot.Spec.Pvc = append(phpBoot.Spec.Pvc, pvcObject)
						},
						Check: func() {
							err := operatorFramework.CreateBootWithError(phpBoot)
							Expect(err).Should(HaveOccurred())
							errStr := fmt.Sprintf("admission webhook \"validation.app.logancloud.com\" denied the request: the pvc %s's label don't match the boot %s", expectPvcName, bootKey.Name)
							Expect(true).Should(Equal(strings.Contains(err.Error(), errStr)))
						},
					})).Run()
				})
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

	Describe("testing validating webhook with env and secret", func() {
		var secret *corev1.Secret
		BeforeEach(func() {
			secret = operatorFramework.SampleSecret(bootKey)
			operatorFramework.CreateSecret(secret)
		})

		Context("test validating webhook with env and secret is OK", func() {
			It("create boot validating webhook with env and secret is OK", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url",
						},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(reflect.DeepEqual(env, envVar)).Should(Equal(true))
							}

							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
					Update: func() {
						boot := operatorFramework.GetBoot(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								env.ValueFrom.SecretKeyRef.Key = "password"
							}
						}
						operatorFramework.UpdateBoot(boot)
					},
					Recheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(env.ValueFrom.SecretKeyRef.Key).Should(Equal("password"))
							}

							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
				})).Run()
			})

			It("update boot validating webhook with env use value is OK", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url",
						},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(reflect.DeepEqual(env, envVar)).Should(Equal(true))
							}

							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
					Update: func() {
						boot := operatorFramework.GetBoot(bootKey)
						updateEnv := make([]corev1.EnvVar, 0)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								env.ValueFrom = nil
								env.Value = "enva"
							}

							updateEnv = append(updateEnv, env)
						}
						boot.Spec.Env = updateEnv
						operatorFramework.UpdateBoot(boot)
					},
					Recheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(env.Value).Should(Equal("enva"))
							}
							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
				})).Run()
			})

			It("update boot validating webhook with env use valueFrom is OK", func() {
				envVar := corev1.EnvVar{
					Name:  "ENVA",
					Value: "ENVA",
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						operatorFramework.CreateBoot(javaBoot)
					},
					Check: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(reflect.DeepEqual(env, envVar)).Should(Equal(true))
							}

							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
					Update: func() {
						boot := operatorFramework.GetBoot(bootKey)
						updateEnv := make([]corev1.EnvVar, 0)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								env.ValueFrom = &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: bootKey.Name,
										},
										Key: "url",
									},
								}
								env.Value = ""
							}

							updateEnv = append(updateEnv, env)
						}
						boot.Spec.Env = updateEnv
						operatorFramework.UpdateBoot(boot)
					},
					Recheck: func() {
						boot := operatorFramework.GetBoot(bootKey)
						Expect(boot.Name).Should(Equal(bootKey.Name))
						deploy := operatorFramework.GetDeployment(bootKey)
						for _, env := range boot.Spec.Env {
							if env.Name == envVar.Name {
								Expect(env.ValueFrom.SecretKeyRef.Key).Should(Equal("url"))
							}
							for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
								if strings.EqualFold(env.Name, j.Name) {
									Expect(env).Should(Equal(j))
								}
							}
						}
					},
				})).Run()
			})
		})

		Context("test validating webhook with env and secret is Error", func() {
			It("env set both value and value from", func() {
				envVar := corev1.EnvVar{
					Name:  "ENVA",
					Value: "val",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url",
						},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("env set value from with Secret and Configmap ", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url",
						},
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("env set value from with Secret Not Found ", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "ENVA",
							},
							Key: "url",
						},
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("env set value from with Secret key not found ", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url_url",
						},
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})

			It("env set value from with Secret has not permission", func() {
				envVar := corev1.EnvVar{
					Name: "ENVA",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bootKey.Name,
							},
							Key: "url",
						},
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{},
					},
				}

				(&(operatorFramework.E2E{
					Build: func() {
						javaBoot.Spec.Env = append(javaBoot.Spec.Env, envVar)
						javaBoot.Name = "new_boot"
						err := operatorFramework.CreateBootWithError(javaBoot)
						Expect(err).Should(HaveOccurred())
					},
				})).Run()
			})
		})
	})
})
