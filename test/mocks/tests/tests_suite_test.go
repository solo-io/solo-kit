package tests

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	T *testing.T
)

func TestTests(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}
