package e2e

import (
	"github.com/logancloud/logan-app-operator/pkg/logan/util/keys"
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("Testing Metric", func() {
	It("Test Metric Service", func() {
		operatorNN := types.NamespacedName{
			Name:      "logan-app-operator",
			Namespace: "logan",
		}

		svr := operatorFramework.GetService(operatorNN)

		Expect(svr.Spec.Ports[0].Port).Should(Equal(int32(8383)))
		Expect(svr.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt(8383)))
		Expect(svr.Spec.Ports[0].Name).Should(Equal("metrics"))

		Expect(svr.Annotations[keys.PrometheusPathAnnotationKey]).Should(Equal(keys.PrometheusPathAnnotationValue))
		Expect(svr.Annotations[keys.PrometheusPortAnnotationKey]).Should(Equal("8383"))
		Expect(svr.Annotations[keys.PrometheusSchemeAnnotationKey]).Should(Equal(keys.PrometheusSchemeAnnotationValue))
		Expect(svr.Annotations[keys.PrometheusScrapeAnnotationKey]).Should(Equal(keys.PrometheusScrapeAnnotationValue))
	})
})
