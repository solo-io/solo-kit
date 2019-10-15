package multicluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSubclients(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Subclients Suite")
}
