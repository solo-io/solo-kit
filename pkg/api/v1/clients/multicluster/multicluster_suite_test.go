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
	kubehelpers "github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/solo-kit/test/helpers"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestMulticluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multicluster Suite")
}

var (
	localLock, remoteLock     *clusterlock.TestClusterLocker
	localClient, remoteClient kubernetes.Interface
	namespace                 string

	_ = SynchronizedBeforeSuite(func() []byte {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}

		// TODO joekelley build out more robust / less redundant multicluster setup and teardown

		localClient = helpers.MustKubeClient()
		remoteClient = remoteKubeClient()
		var err error

		// Acquire locks
		idPrefix := GinkgoRandomSeed()
		localLock, err = clusterlock.NewKubeClusterLocker(localClient, clusterlock.Options{
			IdPrefix: string(idPrefix),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(localLock.AcquireLock()).NotTo(HaveOccurred())
		remoteLock, err = clusterlock.NewKubeClusterLocker(remoteClient, clusterlock.Options{
			IdPrefix: string(idPrefix),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(remoteLock.AcquireLock()).NotTo(HaveOccurred())

		// Create namespaces
		namespace = helpers.RandString(6)
		err = kubeutils.CreateNamespacesInParallel(localClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.CreateNamespacesInParallel(remoteClient, namespace)
		Expect(err).NotTo(HaveOccurred())

		return nil
	}, func([]byte) {})

	_ = SynchronizedAfterSuite(func() {}, func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}
		// Delete namespaces
		err := kubeutils.DeleteNamespacesInParallelBlocking(localClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		err = kubeutils.DeleteNamespacesInParallelBlocking(remoteClient, namespace)
		Expect(err).NotTo(HaveOccurred())
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		// Delete CRDs
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

		// Delete namespaces
		kubehelpers.WaitForNamespaceTeardownWithClient(namespace, localClient)
		kubehelpers.WaitForNamespaceTeardownWithClient(namespace, remoteClient)

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
