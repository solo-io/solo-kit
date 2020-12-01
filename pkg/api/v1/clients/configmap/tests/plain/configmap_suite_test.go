package plain_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/k8s-utils/testutils/clusterlock"
	"github.com/solo-io/solo-kit/test/helpers"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestConfigmap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "plain Configmap Suite")
}

var (
	lock *clusterlock.TestClusterLocker

	_ = SynchronizedBeforeSuite(func() []byte {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}
		kubeClient := helpers.MustKubeClient()
		var err error
		lock, err = clusterlock.NewKubeClusterLocker(kubeClient, clusterlock.Options{
			IdPrefix: fmt.Sprintf("%d", GinkgoRandomSeed()),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		return nil
	}, func([]byte) {})

	_ = SynchronizedAfterSuite(func() {}, func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})
)
