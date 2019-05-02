package util

import (
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// call this function with a var _ = in the test suite file
// tests will panic if there's a previously defined BeforeSuite or AfterSuite
func LockingSuite() bool {
	var (
		lock       *clusterlock.TestClusterLocker
		kubeClient kubernetes.Interface
	)

	var _ = BeforeEach(func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kubeClient, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		lock, err = clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{})
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
	})

	var _ = AfterEach(func() {
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

}
