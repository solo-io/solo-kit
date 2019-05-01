package multicluster_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/errors"
	. "github.com/solo-io/solo-kit/pkg/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/secretconverter"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("WatchKubeconfigs", func() {
	var (
		namespace string
		kube kubernetes.Interface
	)
	BeforeEach(func() {
		namespace = "kubeconfighandler" + testutils.RandString(6)
		kube = helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("returns a channel of kubeconfigs", func() {
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

		kubeConfigs, errs, err := WatchKubeConfigs(context.TODO(), kubeClient, kubeCache)
		Expect(err).NotTo(HaveOccurred())

		var allKubeConfigs KubeConfigs
		Eventually(func() (KubeConfigs, error) {
			select {
			case kcs := <-kubeConfigs:
				allKubeConfigs = kcs
				return kcs, nil
			case err := <-errs:
				return nil, err
			case <-time.After(time.Second * 5):
				return nil, errors.Errorf("timed out waiting for next kubeconfigs snapshot")
			}
		}, time.Minute).Should(HaveLen(2))

		readKc1 := allKubeConfigs[ClusterId(kubeCfg1.Metadata.Ref())].KubeConfig.Config
		readKc2 := allKubeConfigs[ClusterId(kubeCfg2.Metadata.Ref())].KubeConfig.Config
		Expect(readKc1.Clusters).To(Equal(kubeCfg1.KubeConfig.Config.Clusters))
		Expect(readKc2.Clusters).To(Equal(kubeCfg2.KubeConfig.Config.Clusters))
	})
})
