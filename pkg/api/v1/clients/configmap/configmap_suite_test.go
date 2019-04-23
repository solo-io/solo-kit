package configmap_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestConfigmap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configmap Suite")
}
