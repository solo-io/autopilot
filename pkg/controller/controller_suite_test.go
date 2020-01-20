package controller_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	if os.Getenv("RUN_MULTICLUSTER_TESTS") != "1" {
		return
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}
