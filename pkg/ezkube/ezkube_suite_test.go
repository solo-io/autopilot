package ezkube_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEzkube(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ezkube Suite")
}
