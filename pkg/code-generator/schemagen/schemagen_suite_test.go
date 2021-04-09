package schemagen_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSchemaGen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SchemaGen Suite")
}
