package protodep_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProtodep(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protodep Suite")
}
