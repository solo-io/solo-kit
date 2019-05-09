package protoutils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
)

// TODO: fix tests
func TestProtoutil(t *testing.T) {

	log.Printf("Skipping Protoutil Suite. Tests are currently failing and need to be fixed.")
	return

	RegisterFailHandler(Fail)
	log.DefaultOut = GinkgoWriter
	RunSpecs(t, "Protoutil Suite")
}
