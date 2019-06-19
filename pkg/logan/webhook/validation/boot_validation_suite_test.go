package validation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestJavaBoots(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Validation Suite")
}
