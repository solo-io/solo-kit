package statusutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStatusutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Statusutils Suite")
}
