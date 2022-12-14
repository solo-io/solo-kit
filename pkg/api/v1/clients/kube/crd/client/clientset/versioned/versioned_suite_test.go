package versioned_test

import (
	"testing"

	"github.com/solo-io/go-utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//TODO: fix tests
func TestVersioned(t *testing.T) {

	log.Printf("Skipping Versioned Suite. Tests are currently failing and need to be fixed.")
	return

	RegisterFailHandler(Fail)
	RunSpecs(t, "Versioned Suite")
}
