package struct_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubesecretStruct(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubesecretStruct Suite")
}
