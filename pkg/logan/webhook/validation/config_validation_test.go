package validation

import (
	"context"
	"github.com/logancloud/logan-app-operator/pkg/apis"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var _ = Describe("validation config webhook", func() {
	logger := logf.Log.WithName("webhook")
	BeforeEach(func() {
		logf.SetLogger(logf.ZapLoggerTo(os.Stderr, true)) //Debug Output
		err := config.NewConfigFromString(getConfigText())
		if err != nil {
			logger.Error(err, "")
		}

		testenv = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deploy", "crds")},
		}
		err = apis.AddToScheme(scheme.Scheme)
		if err != nil {
			logger.Error(err, "")
		}

		if cfg, err = testenv.Start(); err != nil {
			logger.Error(err, "")
		}
		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		c = mgr.GetClient()
		decoder = mgr.GetAdmissionDecoder()
		stop = make(chan struct{})
		go func() {
			Expect(mgr.Start(stop)).NotTo(HaveOccurred())
			logger.Info("Stopped Manager")
		}()
	})

	AfterEach(func() {
		close(stop)
		testenv.Stop()
	})

	Describe("validating config webhook", func() {
		It("config.yaml not found", func() {
			ar := createRequest("logan-app-operator-config", "logan",
				"ConfigMap", admissionv1beta1.Update,
				[]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: logan-app-operator-config
data:
  config.yaml1: |-
    ## JavaBoot default config12
`))
			expectConfig(ar, false)
		})

		It("target do not match", func() {
			ar := createRequest("logan-app-operator-config123", "logan",
				"ConfigMap", admissionv1beta1.Update,
				[]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: logan-app-operator-config123
data:
  config.yaml1: |-
    ## JavaBoot default config12
`))
			expectConfig(ar, true)
		})

		It("config.yaml is blank", func() {
			ar := createRequest("logan-app-operator-config", "logan",
				"ConfigMap", admissionv1beta1.Update,
				[]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: logan-app-operator-config
data:
  config.yaml: ""
`))
			expectConfig(ar, false)
		})
	})
})

func getConfigValidator() *ConfigValidator {
	validationHandler := &ConfigValidator{}
	validationHandler.OperatorNamespace = "logan"
	validationHandler.InjectClient(c)
	validationHandler.InjectDecoder(decoder)
	return validationHandler
}

func expectConfig(ar *admissionv1beta1.AdmissionRequest, flag bool) {
	validationHandler := getConfigValidator()
	resp := validationHandler.Handle(context.Background(), types.Request{AdmissionRequest: ar})
	Expect(resp.Response.Allowed).Should(Equal(flag))
}
