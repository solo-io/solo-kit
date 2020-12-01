package testutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/k8s-utils/testutils/clusterlock"
	"github.com/solo-io/k8s-utils/testutils/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// call this function with a var _ = in the test suite file
func LockingSuiteEach() bool {
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

func ErrorNotOccuredOrNotFound(err error) {
	if err != nil && !apierrors.IsNotFound(err) {
		Expect(err).NotTo(HaveOccurred())
	}
}
