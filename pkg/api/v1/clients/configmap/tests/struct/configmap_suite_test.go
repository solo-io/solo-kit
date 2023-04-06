package struct_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigmap(t *testing.T) {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "struct Configmap Suite")
}
