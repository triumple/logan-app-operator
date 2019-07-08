package validation

import (
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var testenv *envtest.Environment
var cfg *rest.Config
var c client.Client
var decoder types.Decoder
var stop chan struct{}

func createRequest(name string, namespace string, kind string, operation admissionv1beta1.Operation, raw []byte) *admissionv1beta1.AdmissionRequest {
	ar := &admissionv1beta1.AdmissionRequest{
		Name:      name,
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
