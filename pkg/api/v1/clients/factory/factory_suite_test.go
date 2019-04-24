package factory_test

import (
	"testing"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Factory Suite")
}

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
