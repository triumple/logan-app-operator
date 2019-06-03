package v1

import (
	"log"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var t *envtest.Environment
var cfg *rest.Config
var c client.Client

//func TestMain(m *testing.M) {
//
//
//	code := m.Run()
//
//
//	os.Exit(code)
//}

func TestAppV1(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "AppV1 Suite")
}

var _ = BeforeSuite(func() {
	t = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "deploy", "crds")},
	}

	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err = t.Start()
	if err != nil {
		log.Fatal(err)
	}

	c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})

	if err != nil {
		log.Fatal(err)
	}
})

var _ = AfterSuite(func() {
	t.Stop()
})
