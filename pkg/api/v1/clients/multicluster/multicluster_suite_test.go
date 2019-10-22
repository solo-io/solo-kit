package multicluster_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/solo-kit/test/helpers"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestMulticluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multicluster Suite")
}

var (
	localLock, remoteLock *clusterlock.TestClusterLocker
	err                   error

	_ = SynchronizedBeforeSuite(func() []byte {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}

		// TODO joekelley build out more robust / less redundant multicluster setup and teardown
		// https://github.com/solo-io/go-utils/issues/325

		// Acquire locks
		idPrefix := GinkgoRandomSeed()
		localLock, err = clusterlock.NewKubeClusterLocker(helpers.MustKubeClient(), clusterlock.Options{
			IdPrefix: string(idPrefix),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(localLock.AcquireLock()).NotTo(HaveOccurred())
		remoteLock, err = clusterlock.NewKubeClusterLocker(remoteKubeClient(), clusterlock.Options{
			IdPrefix: string(idPrefix),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(remoteLock.AcquireLock()).NotTo(HaveOccurred())

		return nil
	}, func([]byte) {})

	_ = SynchronizedAfterSuite(func() {}, func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}

		// Delete CRDs
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		apiextsClientset, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		err = apiextsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("anothermockresources.testing.solo.io", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		cfg, err = kubeutils.GetConfig("", os.Getenv("ALT_CLUSTER_KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred())
		remoteApiextsClientset, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		err = remoteApiextsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("anothermockresources.testing.solo.io", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)

		// Release locks
		Expect(localLock.ReleaseLock()).NotTo(HaveOccurred())
		Expect(remoteLock.ReleaseLock()).NotTo(HaveOccurred())
	})
)

// TODO joekelley update util to take an env var arg kube config
func remoteKubeClient() kubernetes.Interface {
	cfg, err := kubeutils.GetConfig("", os.Getenv("ALT_CLUSTER_KUBECONFIG"))
	Expect(err).NotTo(HaveOccurred())
	client, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	return client
}
