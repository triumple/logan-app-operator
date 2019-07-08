package validation

import (
	"context"
	"github.com/logancloud/logan-app-operator/pkg/apis"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/logancloud/logan-app-operator/pkg/logan/webhook"
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

var _ = Describe("validation webhook", func() {
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

	Describe("validating webhook", func() {
		It("good one", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
`))
			expect(ar, true)
		})

		It("ignore namespace ", func() {
			ar := createRequest("default-javaboot", "logan-dev",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`
apiVersison: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
`))
			expect(ar, true)
		})

		It("empty raw", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(``))
			expect(ar, true)
		})

		It("error decode ", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`{"sds":"sds"""}`))
			expect(ar, false)
		})

		It("unknow kind ", func() {
			ar := createRequest("default-javaboot", "logan",
				"unknow", admissionv1beta1.Create,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
`))
			expect(ar, false)
		})

		It("unknow kind in raw ", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: unknow
metadata:
  name: default-javaboot
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
`))
			expect(ar, false)
		})

		It("check env with create operation", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: SPRING_ZIPKIN_ENABLED
      value: "false"
`))
			expect(ar, false)
		})

		It("check env with update operation ", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Update,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
  annotations:
    app.logancloud.com/boot-envs: >-
      [{"name":"B","value":"B"},{"name":"A","value":"A"}]

spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: B
      value: "new_b"
    - name: C
      value: "C"
`))
			expect(ar, true)
		})

		It("check env with update operation and empty env annotations", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Update,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
  annotations:
    app.logancloud.com/boot-envs:
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: B
      value: "b"
    - name: C
      value: "C"
`))
			expect(ar, true)
		})

		It("delete global env with update operation ", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Update,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
  annotations:
    app.logancloud.com/boot-envs: >-
      [{"name":"SPRING_ZIPKIN_ENABLED","value":"true"},{"name":"B","value":"B"}]

spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: B
      value: "new_b"
    - name: C
      value: "C"
`))
			expect(ar, false)
		})

		It("add global env with update operation ", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Update,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
  annotations:
    app.logancloud.com/boot-envs: >-
      [{"name":"B","value":"B"}]

spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: B
      value: "new_b"
    - name: SPRING_ZIPKIN_ENABLED
      value: "false"
`))

			expect(ar, false)
		})

		It("modify global env with update operation", func() {
			ar := createRequest("default-javaboot", "logan",
				webhook.ApiTypeJava, admissionv1beta1.Update,
				[]byte(`
apiVersion: app.logancloud.com/v1
kind: JavaBoot
metadata:
  name: default-javaboot
  annotations:
    app.logancloud.com/boot-envs: >-
      [{"name":"SPRING_ZIPKIN_ENABLED","value":"true"},{"name":"B","value":"B"}]
spec:
  replicas: 1
  image: "logan-startkit-boot"
  version: "1.2.1"
  env:
    - name: B
      value: "new_b"
    - name: SPRING_ZIPKIN_ENABLED
      value: "false"
`))
			expect(ar, false)
		})
	})
})

func getBootValidator() *BootValidator {
	validationHandler := &BootValidator{}
	validationHandler.InjectClient(c)
	validationHandler.InjectDecoder(decoder)
	return validationHandler
}

func expect(ar *admissionv1beta1.AdmissionRequest, flag bool) {
	validationHandler := getBootValidator()
	resp := validationHandler.Handle(context.Background(), types.Request{AdmissionRequest: ar})
	Expect(resp.Response.Allowed).Should(Equal(flag))
}
