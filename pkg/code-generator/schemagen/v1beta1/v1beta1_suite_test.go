package v1beta1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV1Beta1SchemaGen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v1beta1 SchemaGen Suite")
}
