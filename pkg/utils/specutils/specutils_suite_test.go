package specutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Specutils Suite")
}
