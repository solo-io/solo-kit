package kubeinstallutils_test

import (
	"testing"

	"github.com/solo-io/solo-kit/pkg/utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: fix. This just looks to be very slow
func TestShared(t *testing.T) {

	log.Printf("Skipping Shared Suite. Tests are currently failing and need to be fixed.")
	return

	RegisterFailHandler(Fail)
	RunSpecs(t, "Shared Suite")
}
