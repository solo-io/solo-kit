package multicluster_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	. "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	"github.com/solo-io/solo-kit/test/setup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("KubeConfigHandler", func() {
	var (
		namespace string
	)
	BeforeEach(func() {
		namespace = "kubeconfighandler" + testutils.RandString(6)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("calls a callback for watched kube configs", func() {
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kubeClient, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		kubeConfig, err := kubeutils.GetKubeConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kubeCfgSecret1, err := secretconverter.KubeConfigToSecret(v1.ObjectMeta{Name: "kubeconfig1", Namespace: namespace}, kubeConfig)
		Expect(err).NotTo(HaveOccurred())
		kubeCfgSecret1, err = kubeClient.CoreV1().Secrets(namespace).Create(kubeCfgSecret1)
		Expect(err).NotTo(HaveOccurred())
		kubeCfg1, err := secretconverter.KubeCfgFromSecret(kubeCfgSecret1)
		Expect(err).NotTo(HaveOccurred())

		kubeCfgSecret2, err := secretconverter.KubeConfigToSecret(v1.ObjectMeta{Name: "kubeconfig2", Namespace: namespace}, kubeConfig)
		Expect(err).NotTo(HaveOccurred())
		kubeCfgSecret2, err = kubeClient.CoreV1().Secrets(namespace).Create(kubeCfgSecret2)
		Expect(err).NotTo(HaveOccurred())
		kubeCfg2, err := secretconverter.KubeCfgFromSecret(kubeCfgSecret2)
		Expect(err).NotTo(HaveOccurred())

		kubeCache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
		Expect(err).NotTo(HaveOccurred())
		rch, err := NewKubeConfigHandler(kubeClient, kubeCache)
		Expect(err).NotTo(HaveOccurred())

		var allKubeConfigs KubeConfigs
		rch.SetCallback(func(updated KubeConfigs) error {
			allKubeConfigs = updated
			return nil
		})

		errs, err := rch.Start(context.TODO())
		Expect(err).NotTo(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			err := <-errs
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() KubeConfigs {
			return allKubeConfigs
		}, time.Minute).Should(HaveLen(2))


		readKc1 := allKubeConfigs[ClusterId(kubeCfg1.Metadata.Name)].KubeConfig.KubeConfig
		readKc2 := allKubeConfigs[ClusterId(kubeCfg2.Metadata.Name)].KubeConfig.KubeConfig
		Expect(readKc1.Clusters).To(Equal(kubeCfg1.KubeConfig.KubeConfig.Clusters))
		Expect(readKc2.Clusters).To(Equal(kubeCfg2.KubeConfig.KubeConfig.Clusters))
	})
})
