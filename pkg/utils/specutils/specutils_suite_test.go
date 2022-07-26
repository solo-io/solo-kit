package specutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSpecutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Specutils Suite")
}
