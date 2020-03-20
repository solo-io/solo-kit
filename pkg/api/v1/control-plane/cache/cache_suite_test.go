package cache

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestControlPlaneCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControlPlaneCache Suite")
}
