package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestNamespace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configmap Suite")
}
