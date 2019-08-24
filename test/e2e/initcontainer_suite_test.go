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

var _ = Describe("Test initContainer", func() {
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

	Describe("Testing InitContainer [Serial]", func() {
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

		It("testing InitContainer create and update", func() {
			(&(operatorFramework.E2E{
				Build: func() {
					operatorFramework.CreateBoot(phpBoot)
				},
				Check: func() {
					boot := operatorFramework.GetPhpBoot(bootKey)
					Expect(boot.Name).Should(Equal(bootKey.Name))

					deploy := operatorFramework.GetDeployment(bootKey)
					hasInitContainer := false
					isEnvTestKeyExisted := false
					for _, container := range deploy.Spec.Template.Spec.InitContainers {
						if container.Name == "fetcher" {
							hasInitContainer = true
							for _, v := range container.Env {
								if v.Name == "APPLICATION_NAME" {
									isEnvTestKeyExisted = true
									Expect(v.Value).Should(Equal(bootKey.Name))
								}
							}
						}
					}
					Expect(hasInitContainer).Should(Equal(true))
					Expect(isEnvTestKeyExisted).Should(Equal(true))
				},
				Update: func() {
					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]
					((*operator).AppSpec.PodSpec.InitContainers)[0].Env = append(((*operator).AppSpec.PodSpec.InitContainers)[0].Env, env)
					c[logan.BootPhp] = operator

					updatedInitContainerContent, _ := ghodssyaml.Marshal(&c)

					data := make(map[string]string)
					data["config.yaml"] = string(updatedInitContainerContent)
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
					hasInitContainer := false
					isEnvTestKeyExisted := false
					for _, container := range deploy.Spec.Template.Spec.InitContainers {
						if container.Name == "fetcher" {
							hasInitContainer = true
							for _, v := range container.Env {
								if v.Name == env.Name {
									isEnvTestKeyExisted = true
									Expect(v.Value).Should(Equal(env.Value))
								}
							}
						}
					}
					Expect(hasInitContainer).Should(Equal(true))
					Expect(isEnvTestKeyExisted).Should(Equal(true))
				},
			})).Run()
		})

		It("testing InitContainer env override", func() {
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
					hasInitContainer := false
					for _, container := range deploy.Spec.Template.Spec.InitContainers {
						if container.Name == "fetcher" {
							hasInitContainer = true
						}
					}
					Expect(hasInitContainer).Should(Equal(true))
				},
				Update: func() {
					c := operatorFramework.GetConfig(configNN)
					operator := c[logan.BootPhp]
					((*operator).AppSpec.PodSpec.InitContainers)[0].Env = append(((*operator).AppSpec.PodSpec.InitContainers)[0].Env, env)

					initContainerEnvs, ok := operator.OEnvs["fetcher"]
					if !ok {
						initContainerEnvs = make(map[string]config.AppSpec)
					}
					initContainerEnv, ok := initContainerEnvs["test"]
					if !ok {
						var appSpec config.AppSpec
						initContainerEnv = appSpec
					}
					initContainerEnv.Env = append(initContainerEnv.Env, replaceEnv)
					initContainerEnvs["test"] = initContainerEnv
					operator.OEnvs["fetcher"] = initContainerEnvs

					c[logan.BootPhp] = operator

					updatedInitContainerContent, _ := ghodssyaml.Marshal(&c)

					data := make(map[string]string)
					data["config.yaml"] = string(updatedInitContainerContent)
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
					hasInitContainer := false
					isEnvTestKeyExisted := false
					for _, container := range deploy.Spec.Template.Spec.InitContainers {
						if container.Name == "fetcher" {
							hasInitContainer = true
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
					Expect(hasInitContainer).Should(Equal(true))
					Expect(isEnvTestKeyExisted).Should(Equal(true))
				},
			})).Run()
		})
	})
})
