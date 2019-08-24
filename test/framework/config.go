package framework

import (
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

func GetConfigStr(configNN types.NamespacedName) string {
	configMap := GetConfigmap(configNN)
	configContent := ""
	for k, v := range configMap.Data {
		if k == "config.yaml" {
			configContent = v
			break
		}
	}

	return configContent
}

func GetConfig(configNN types.NamespacedName) config.GlobalConfig {
	configContent := GetConfigStr(configNN)

	Expect(configContent).ShouldNot(Equal(""))
	content := ioutil.NopCloser(strings.NewReader(configContent))
	c := config.GlobalConfig{}

	err := k8syaml.NewYAMLOrJSONDecoder(content, 100).Decode(&c)
	Expect(err).ShouldNot(HaveOccurred())

	return c
}
