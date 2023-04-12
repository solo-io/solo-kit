package tests

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
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
