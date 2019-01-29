package code_generator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCodeGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CodeGenerator Suite")
}
