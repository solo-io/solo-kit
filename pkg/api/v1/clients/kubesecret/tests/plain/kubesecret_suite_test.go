package plain_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubesecretPlain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubesecretPlain Suite")
}
