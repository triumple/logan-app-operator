package validation

import (
	"context"
	"github.com/logancloud/logan-app-operator/pkg/apis"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/logancloud/logan-app-operator/pkg/logan/webhook"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var testenv *envtest.Environment
var cfg *rest.Config
var c client.Client
var decoder types.Decoder
var stop chan struct{}

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
			ar := createRequest("logan",
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
			ar := createRequest("logan-dev",
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
			ar := createRequest("logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(``))
			expect(ar, true)
		})

		It("error decode ", func() {
			ar := createRequest("logan",
				webhook.ApiTypeJava, admissionv1beta1.Create,
				[]byte(`{"sds":"sds"""}`))
			expect(ar, false)
		})

		It("unknow kind ", func() {
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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
			ar := createRequest("logan",
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

func expect(ar *admissionv1beta1.AdmissionRequest, flag bool) {
	validationHandler := getBootValidator()
	resp := validationHandler.Handle(context.Background(), types.Request{AdmissionRequest: ar})
	Expect(resp.Response.Allowed).Should(Equal(flag))
}

func createRequest(namespace string, kind string, operation admissionv1beta1.Operation, raw []byte) *admissionv1beta1.AdmissionRequest {
	ar := &admissionv1beta1.AdmissionRequest{
		Namespace: namespace,
		Operation: operation,
		Kind: metav1.GroupVersionKind{
			Kind: kind,
		},
		Object: runtime.RawExtension{
			Raw: raw,
		},
	}
	return ar
}

func getBootValidator() *BootValidator {
	validationHandler := &BootValidator{}
	validationHandler.InjectClient(c)
	validationHandler.InjectDecoder(decoder)
	return validationHandler
}

func getConfigText() string {
	configText := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
        port: 8082
        replicas: 2
        health: /health2
        env:
          # Podpreset
          - name: SPRING_ZIPKIN_ENABLED2
            value: "false"
        nodeSelector:
          logan/envA: A
          logan/envB: B
        resources:
          limits:
            cpu: "4"
            memory: "4Gi"
          requests:
            cpu: "3"
            memory: "3Gi"
        subDomain: "2exp.logan.local"
  app:
    port: 8083
    replicas: 3
    health: /health3
    env:
      # Podpreset
      - name: SPRING_ZIPKIN_ENABLED
        value: "true"
    nodeSelector:
      logan/envA: NewA
      logan/envC: C
    resources:
      limits:
        cpu: "2"
        memory: "2Gi"
      requests:
        cpu: "1"
        memory: "1Gi"
    subDomain: "3exp.logan.local"
`
	return configText
}
