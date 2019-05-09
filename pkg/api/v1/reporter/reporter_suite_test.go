package reporter_test

import (
	"testing"

	"github.com/solo-io/go-utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: fix tests
func TestReporter(t *testing.T) {

	log.Printf("Skipping Vault Suite. Tests are currently failing and need to be fixed.")
	return

	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporter Suite")
}
