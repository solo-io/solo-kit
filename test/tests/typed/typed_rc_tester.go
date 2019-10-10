package typed

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/test/helpers"
	"k8s.io/client-go/rest"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"

	"github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/setup"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type ResourceClientTester interface {
	Description() string
	Skip() bool
	Setup(namespace string) factory.ResourceClientFactory
	Teardown(namespace string)
	// NOTE: ResourceClientFactoryGetter functions will not work until setup has been called
	factory.ResourceClientFactoryGetter
}

func skipKubeTests() bool {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return true
	}
	return false
}

/*
 *
 * KubeCrd
 *
 */

type KubeRcTester struct {
	Crd crd.Crd
}

func (rct *KubeRcTester) Description() string {
	return "kube-crd"
}

func (rct *KubeRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeRcTester) Setup(namespace string) factory.ResourceClientFactory {
	if namespace != "" {
		kubeClient := helpers.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
	}
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeResourceClientFactory{
		Crd:         rct.Crd,
		Cfg:         cfg,
		SharedCache: kube.NewKubeCache(context.TODO()),
	}
}

func (rct *KubeRcTester) Teardown(namespace string) {
	kubeClient := helpers.MustKubeClient()
	err := kubeutils.DeleteNamespacesInParallelBlocking(kubeClient, namespace)
	Expect(err).NotTo(HaveOccurred())
}

func (rct *KubeRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	return &factory.KubeResourceClientFactory{
		Crd:         rct.Crd,
		Cfg:         restConfig,
		SharedCache: kube.NewKubeCache(context.TODO()),
		Cluster:     cluster,
	}
}

/*
 *
 * Consul-KV
 *
 */
type ConsulRcTester struct {
	consulInstance *setup.ConsulInstance
	consulFactory  *setup.ConsulFactory
	consul         *api.Client
	namespace      string
}

func (rct *ConsulRcTester) Description() string {
	return "consul-kv"
}

func (rct *ConsulRcTester) Skip() bool {
	if os.Getenv("RUN_CONSUL_TESTS") != "1" {
		log.Printf("This test downloads and runs consul and is disabled by default. To enable, set RUN_CONSUL_TESTS=1 in your env.")
		return true
	}
	return false
}

func (rct *ConsulRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.consulFactory, err = setup.NewConsulFactory()
	Expect(err).NotTo(HaveOccurred())
	rct.consulInstance, err = rct.consulFactory.NewConsulInstance()
	Expect(err).NotTo(HaveOccurred())
	err = rct.consulInstance.Run()
	Expect(err).NotTo(HaveOccurred())

	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("127.0.0.1:%v", rct.consulInstance.Ports.HttpPort)
	consul, err := api.NewClient(cfg)
	Expect(err).NotTo(HaveOccurred())

	rct.consul = consul
	rct.namespace = namespace
	return &factory.ConsulResourceClientFactory{
		Consul:  consul,
		RootKey: namespace,
	}
}

func (rct *ConsulRcTester) Teardown(namespace string) {
	rct.consulInstance.Clean()
	rct.consulFactory.Clean()
}

func (rct *ConsulRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	return &factory.ConsulResourceClientFactory{
		Consul:  rct.consul,
		RootKey: rct.namespace,
	}
}

/*
 *
 * File
 *
 */
type FileRcTester struct {
	rootDir string
}

func (rct *FileRcTester) Description() string {
	return "file-based"
}

func (rct *FileRcTester) Skip() bool {
	return false
}

func (rct *FileRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.rootDir, err = ioutil.TempDir("", "base_test")
	Expect(err).NotTo(HaveOccurred())
	return &factory.FileResourceClientFactory{
		RootDir: rct.rootDir,
	}
}

func (rct *FileRcTester) Teardown(namespace string) {
	os.RemoveAll(rct.rootDir)
}

func (rct *FileRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	return &factory.FileResourceClientFactory{
		RootDir: rct.rootDir,
	}
}

/*
 *
 * Memory
 *
 */
type MemoryRcTester struct {
	rootDir string
}

func (rct *MemoryRcTester) Description() string {
	return "memory-based"
}

func (rct *MemoryRcTester) Skip() bool {
	return false
}

func (rct *MemoryRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.rootDir, err = ioutil.TempDir("", "base_test")
	Expect(err).NotTo(HaveOccurred())
	return &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

func (rct *MemoryRcTester) Teardown(namespace string) {}

func (rct *MemoryRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	return &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

/*
 *
 * KubeCfgMap
 *
 */
type KubeConfigMapRcTester struct{}

func (rct *KubeConfigMapRcTester) Description() string {
	return "kube-configmap-based"
}

func (rct *KubeConfigMapRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeConfigMapRcTester) Setup(namespace string) factory.ResourceClientFactory {
	kubeClient := helpers.MustKubeClient()
	err := kubeutils.CreateNamespacesInParallel(kubeClient, namespace)
	Expect(err).NotTo(HaveOccurred())
	kcache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeConfigMapClientFactory{
		Clientset: kubeClient,
		Cache:     kcache,
	}
}

func (rct *KubeConfigMapRcTester) Teardown(namespace string) {
	kubeClient := helpers.MustKubeClient()
	err := kubeutils.DeleteNamespacesInParallelBlocking(kubeClient, namespace)
	Expect(err).NotTo(HaveOccurred())
}

func (rct *KubeConfigMapRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	kubeClient := helpers.MustKubeClient()
	kcache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeConfigMapClientFactory{
		Clientset: kubeClient,
		Cache:     kcache,
	}
}

/*
 *
 * KubeSecret
 *
 */
type KubeSecretRcTester struct{}

func (rct *KubeSecretRcTester) Description() string {
	return "kube-secret-based"
}

func (rct *KubeSecretRcTester) Skip() bool {
	return skipKubeTests()
}

func (rct *KubeSecretRcTester) Setup(namespace string) factory.ResourceClientFactory {
	kubeClient := helpers.MustKubeClient()
	err := kubeutils.CreateNamespacesInParallel(kubeClient, namespace)
	Expect(err).NotTo(HaveOccurred())
	kcache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeSecretClientFactory{
		Clientset: kubeClient,
		Cache:     kcache,
	}
}

func (rct *KubeSecretRcTester) Teardown(namespace string) {
	kubeClient := helpers.MustKubeClient()
	err := kubeutils.DeleteNamespacesInParallelBlocking(kubeClient, namespace)
	Expect(err).NotTo(HaveOccurred())
}

func (rct *KubeSecretRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	kubeClient := helpers.MustKubeClient()
	kcache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	Expect(err).NotTo(HaveOccurred())
	return &factory.KubeSecretClientFactory{
		Clientset: kubeClient,
		Cache:     kcache,
	}
}

/*
 *
 * Vault Secret
 *
 */
type VaultRcTester struct {
	vaultInstance *setup.VaultInstance
	vaultFactory  *setup.VaultFactory
	rootKey       string
	vault         *vaultapi.Client
}

func (rct *VaultRcTester) Description() string {
	return "vault-secret-based"
}

func (rct *VaultRcTester) Skip() bool {
	if os.Getenv("RUN_VAULT_TESTS") != "1" {
		log.Printf("This test downloads and runs vault and is disabled by default. To enable, set RUN_VAULT_TESTS=1 in your env.")
		return true
	}
	return false
}

func (rct *VaultRcTester) Setup(namespace string) factory.ResourceClientFactory {
	var err error
	rct.vaultFactory, err = setup.NewVaultFactory()
	Expect(err).NotTo(HaveOccurred())
	rct.vaultInstance, err = rct.vaultFactory.NewVaultInstance()
	Expect(err).NotTo(HaveOccurred())
	err = rct.vaultInstance.Run()
	Expect(err).NotTo(HaveOccurred())
	rootKey := "/secret/" + namespace
	cfg := vaultapi.DefaultConfig()
	cfg.Address = fmt.Sprintf("http://127.0.0.1:%v", rct.vaultInstance.Port)
	vault, err := vaultapi.NewClient(cfg)
	vault.SetToken(rct.vaultInstance.Token())
	Expect(err).NotTo(HaveOccurred())
	rct.rootKey = rootKey
	rct.vault = vault
	return &factory.VaultSecretClientFactory{
		RootKey: rootKey,
		Vault:   vault,
	}
}

func (rct *VaultRcTester) Teardown(namespace string) {
	rct.vaultInstance.Clean()
	rct.vaultFactory.Clean()
}

func (rct *VaultRcTester) ForCluster(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
	return &factory.VaultSecretClientFactory{
		RootKey: rct.rootKey,
		Vault:   rct.vault,
	}
}
