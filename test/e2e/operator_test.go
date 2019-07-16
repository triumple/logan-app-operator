package e2e

import (
	operatorFramework "github.com/logancloud/logan-app-operator/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"os"
	"testing"
)

var (
	framework *operatorFramework.Framework
)

func TestMain(m *testing.M) {
	var (
		err      error
		exitCode int
	)
	if framework, err = operatorFramework.InitFramework(); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}
	exitCode = m.Run()
	os.Exit(exitCode)
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Operator Suite")
}
