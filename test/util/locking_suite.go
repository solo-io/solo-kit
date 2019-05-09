package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/go-utils/testutils/kube"
)

// call this function with a var _ = in the test suite file
// tests will panic if there's a previously defined BeforeSuite or AfterSuite
func LockingSuite() bool {
	var (
		lock *clusterlock.TestClusterLocker
	)

	var _ = BeforeEach(func() {
		var err error
		lock, err = clusterlock.NewTestClusterLocker(kube.MustKubeClient(), clusterlock.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
	})

	var _ = AfterEach(func() {
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	return true
}
