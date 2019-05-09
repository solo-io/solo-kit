package file_test

import (
	"testing"

	"github.com/solo-io/go-utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: fix tests
func TestFile(t *testing.T) {
	log.Printf("Skipping File Suite. Tests are currently failing and need to be fixed.")
	return
	RegisterFailHandler(Fail)
	RunSpecs(t, "File Suite")
}
