package sk_anyvendor_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSkAnyvendor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SkAnyvendor Suite")
}
