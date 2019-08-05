package e2e

import (
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
)

var _ = Describe("Test sidecar", func() {
	var bootKey types.NamespacedName
	var phpBoot *bootv1.PhpBoot
	var configNN = types.NamespacedName{
		Name:      "logan-app-operator-config",
		Namespace: "logan",
	}
	var env = corev1.EnvVar{
		Name:  "testKey",
		Value: "testValue",
	}

	var envForUpdateBoot = corev1.EnvVar{
		Name:  "A",
		Value: "A",
	}

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

	Describe("Testing sidecar [Serial]", func() {
		var configYamlStr string
		BeforeEach(func() {
			// backup config map: config.yaml
			configYamlStr = operatorFramework.GetConfigStr(configNN)
		})

		AfterEach(func() {
			// recover config map: config.yaml
			data := make(map[string]string)
			data["config.yaml"] = configYamlStr
			var configMap = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: configNN.Name, Namespace: configNN.Namespace},
				Data:       data,
			}
			operatorFramework.UpdateConfigmap(&configMap)
		})

		It("testing sidecar create and update", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(phpBoot)
				},
				Check: func() {
					boot := operatorFramework.GetPhpBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))
					Expect(boot.Spec.Port).Should(Equal(int32(7777)))
					Expect(boot.Spec.Prometheus).Should(Equal("true"))

					deploy := operatorFramework.GetDeployment(bootKey)
					hasSidecarContainer := false
					for _, container := range deploy.Spec.Template.Spec.Containers {
						if container.Name == "sidecar" {
							hasSidecarContainer = true
							Expect(container.Ports[0].ContainerPort).Should(Equal(int32(5678)))
							Expect(container.Ports[0].Name).Should(Equal("http"))

							isEnvKeyExisted := false
							for _, v := range container.Env {
								if v.Name == "SPRING_PROFILES_ACTIVE" {
									isEnvKeyExisted = true
									Expect(v.Value).Should(Equal("test"))
									break
								}
							}
							Expect(isEnvKeyExisted).Should(Equal(true))
						}
					}
					Expect(hasSidecarContainer).Should(Equal(true))

					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]
					sidecarBootKey := types.NamespacedName{
						Name:      bootKey.Name + "-" + (*operator.SidecarContainers)[0].Ports[0].Name,
						Namespace: bootKey.Namespace,
					}
					service := operatorFramework.GetService(sidecarBootKey)
					Expect(service.Name).Should(Equal(sidecarBootKey.Name))
					Expect(service.Spec.Ports[0].Name).Should(Equal("http"))
					Expect(service.Spec.Ports[0].Port).Should(Equal(int32(5678)))
				},
				Update: func() {
					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]

					(*operator.SidecarContainers)[0].Env = append((*operator.SidecarContainers)[0].Env, env)
					c[logan.BootPhp] = operator
					updatedConfigContent, _ := ghodssyaml.Marshal(&c)

					data := make(map[string]string)
					data["config.yaml"] = string(updatedConfigContent)
					var configMap = corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: configNN.Name, Namespace: configNN.Namespace},
						Data:       data,
					}
					operatorFramework.UpdateConfigmap(&configMap)

					phpBoot.Spec.Env = append(phpBoot.Spec.Env, envForUpdateBoot)
					operatorFramework.UpdatePhpBoot(phpBoot)
				},
				Recheck: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					hasSidecarContainer := false
					isEnvTestKeyExisted := false
					for _, container := range deploy.Spec.Template.Spec.Containers {
						if container.Name == "sidecar" {
							hasSidecarContainer = true
							for _, v := range container.Env {
								if v.Name == env.Name {
									isEnvTestKeyExisted = true
									Expect(v.Value).Should(Equal(env.Value))
								}
							}
						}
					}
					Expect(hasSidecarContainer).Should(Equal(true))
					Expect(isEnvTestKeyExisted).Should(Equal(true))
				},
			})).Run()
		})

		It("testing sidecar env override", func() {
			var replaceEnv = corev1.EnvVar{
				Name:  "testKey",
				Value: "replaceTestValue",
			}
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(phpBoot)
				},
				Check: func() {
					boot := operatorFramework.GetPhpBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))

					deploy := operatorFramework.GetDeployment(bootKey)
					hasSidecarContainer := false
					for _, container := range deploy.Spec.Template.Spec.Containers {
						if container.Name == "sidecar" {
							hasSidecarContainer = true
						}
					}
					Expect(hasSidecarContainer).Should(Equal(true))

					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]
					sidecarBootKey := types.NamespacedName{
						Name:      bootKey.Name + "-" + (*operator.SidecarContainers)[0].Ports[0].Name,
						Namespace: bootKey.Namespace,
					}
					service := operatorFramework.GetService(sidecarBootKey)
					Expect(service.Name).Should(Equal(sidecarBootKey.Name))
				},
				Update: func() {
					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]
					(*operator.SidecarContainers)[0].Env = append((*operator.SidecarContainers)[0].Env, env)
					sideCarEnvs, ok := operator.OEnvs["sidecar"]
					if !ok {
						sideCarEnvs = make(map[string]config.AppSpec)
					}
					sideCarEnv, ok := sideCarEnvs["test"]
					if !ok {
						var appSpec config.AppSpec
						sideCarEnv = appSpec
					}
					sideCarEnv.Env = append(sideCarEnv.Env, replaceEnv)
					sideCarEnvs["test"] = sideCarEnv
					operator.OEnvs["sidecar"] = sideCarEnvs

					c[logan.BootPhp] = operator
					updatedConfigContent, _ := ghodssyaml.Marshal(&c)

					data := make(map[string]string)
					data["config.yaml"] = string(updatedConfigContent)
					var configMap = corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: configNN.Name, Namespace: configNN.Namespace},
						Data:       data,
					}
					operatorFramework.UpdateConfigmap(&configMap)

					phpBoot.Spec.Env = append(phpBoot.Spec.Env, envForUpdateBoot)
					operatorFramework.UpdatePhpBoot(phpBoot)
				},
				Recheck: func() {
					deploy := operatorFramework.GetDeployment(bootKey)
					hasSidecarContainer := false
					isEnvTestKeyExisted := false
					for _, container := range deploy.Spec.Template.Spec.Containers {
						if container.Name == "sidecar" {
							hasSidecarContainer = true
							for _, v := range container.Env {
								if v.Name == env.Name {
									isEnvTestKeyExisted = true
									Expect(v.Value).Should(Equal(replaceEnv.Value))
									break
								}
							}
							break
						}
					}
					Expect(hasSidecarContainer).Should(Equal(true))
					Expect(isEnvTestKeyExisted).Should(Equal(true))
				},
			})).Run()
		})
	})
})
