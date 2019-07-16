// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestMulticlustersoloio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multiclustersoloio Suite")
}

var (
	lock *clusterlock.TestClusterLocker
	cfg  *rest.Config

	_ = SynchronizedAfterSuite(func() {}, func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		clientset, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("anothermockresources.testing.solo.io.v1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("clusterresources.testing.solo.io.v1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("fakes.testing.solo.io.v1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("fakes.testing.solo.io.v1alpha1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("mocks.mocking.solo.io.v1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("mocks.testing.solo.io.v1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("mocks.testing.solo.io.v1alpha1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("mocks.testing.solo.io.v2alpha1", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	_ = SynchronizedBeforeSuite(func() []byte {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		clientset, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		lock, err = clusterlock.NewTestClusterLocker(clientset, clusterlock.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		return nil
	}, func([]byte) {})
)
