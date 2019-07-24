package framework

import (
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetConfigmap(nn types.NamespacedName) *v1.ConfigMap {
	configMap := &v1.ConfigMap{}
	var err error
	Eventually(func() error {
		configMap, err = framework.KubeClient.CoreV1().ConfigMaps(nn.Namespace).Get(nn.Name, metav1.GetOptions{})
		return err
	}, defaultTimeout).
		Should(Succeed())
	return configMap
}

func UpdateConfigmap(configMap *v1.ConfigMap) *v1.ConfigMap {
	conf := &v1.ConfigMap{}
	var err error
	Eventually(func() error {
		conf, err = framework.KubeClient.CoreV1().ConfigMaps(configMap.Namespace).Update(configMap)
		return err
	}, defaultTimeout).
		Should(Succeed())
	WaitDefaultUpdate()
	return conf
}

func DeleteConfigmap(configMap *v1.ConfigMap) {
	Eventually(func() error {
		return framework.KubeClient.CoreV1().ConfigMaps(configMap.Namespace).Delete(configMap.Name, &metav1.DeleteOptions{})
	}, defaultTimeout).
		Should(Succeed())
}
