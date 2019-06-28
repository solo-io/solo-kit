package customresourcedefinition_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/test/testutils"
)

func TestCustomresourcedefinition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Customresourcedefinition Suite")
}

var _ = testutils.LockingSuiteEach()
