package framework

import (
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

// GetConfigStr will return operator config map string
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

// GetConfig will return operator GlobalConfig struct object
func GetConfig(configNN types.NamespacedName) config.GlobalConfig {
	configContent := GetConfigStr(configNN)

	gomega.Expect(configContent).ShouldNot(gomega.Equal(""))
	content := ioutil.NopCloser(strings.NewReader(configContent))
	c := config.GlobalConfig{}

	err := k8syaml.NewYAMLOrJSONDecoder(content, 100).Decode(&c)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	return c
}
