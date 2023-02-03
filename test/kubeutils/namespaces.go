package kubeutils

// Duplicate of k8s-utils/kubeutils/namespaces.go

import (
	"context"

	"github.com/onsi/ginkgo/v2"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespacesInParallel(ctx context.Context, kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		ns := namespace
		eg.Go(func() error {
			defer ginkgo.GinkgoRecover()

			_, err := kube.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ns,
				},
			}, metav1.CreateOptions{})
			return err
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallelBlocking(ctx context.Context, kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		ns := namespace
		eg.Go(func() error {
			defer ginkgo.GinkgoRecover()

			return kube.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallel(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
	for _, namespace := range namespaces {
		ns := namespace
		go func() {
			defer ginkgo.GinkgoRecover()

			kube.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		}()
	}
}
