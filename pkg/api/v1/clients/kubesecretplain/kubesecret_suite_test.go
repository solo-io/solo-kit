package kubesecretplain_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubesecret(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubesecretplain Suite")
}
