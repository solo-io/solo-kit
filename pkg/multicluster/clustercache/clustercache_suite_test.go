package clustercache_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClustercache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clustercache Suite")
}
