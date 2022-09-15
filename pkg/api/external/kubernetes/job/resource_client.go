package job

import (
	"context"
	"sort"

	kubejob "github.com/solo-io/solo-kit/api/external/kubernetes/job"
	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type jobResourceClient struct {
	cache cache.KubeJobCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
}

func newResourceClient(kube kubernetes.Interface, cache cache.KubeJobCache) *jobResourceClient {
	return &jobResourceClient{
		cache: cache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: &skkube.Job{},
		},
	}
}

func NewJobClient(kube kubernetes.Interface, cache cache.KubeJobCache) skkube.JobClient {
	resourceClient := newResourceClient(kube, cache)
	return skkube.NewJobClientWithBase(resourceClient)
}

func FromKubeJob(job *batchv1.Job) *skkube.Job {

	jobCopy := job.DeepCopy()
	kubeJob := kubejob.Job(*jobCopy)
	resource := &skkube.Job{
		Job: kubeJob,
	}

	return resource
}

func ToKubeJob(resource resources.Resource) (*batchv1.Job, error) {
	jobResource, ok := resource.(*skkube.Job)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to job-only client", resources.Kind(resource))
	}

	job := batchv1.Job(jobResource.Job)

	return &job, nil
}

var _ clients.ResourceClient = &jobResourceClient{}

func (rc *jobResourceClient) RegisterNamespace(namespace string) error {
	return nil
}

func (rc *jobResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	jobObj, err := rc.Kube.BatchV1().Jobs(namespace).Get(opts.Ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading jobObj from kubernetes")
	}
	resource := FromKubeJob(jobObj)

	if resource == nil {
		return nil, errors.Errorf("jobObj %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *jobResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)
	jobObj, err := ToKubeJob(resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.Kube.BatchV1().Jobs(meta.Namespace).Update(opts.Ctx, jobObj, metav1.UpdateOptions{}); err != nil {
			return nil, errors.Wrapf(err, "updating kube jobObj %v", jobObj.Name)
		}
	} else {
		if _, err := rc.Kube.BatchV1().Jobs(meta.Namespace).Create(opts.Ctx, jobObj, metav1.CreateOptions{}); err != nil {
			return nil, errors.Wrapf(err, "creating kube jobObj %v", jobObj.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(jobObj.Namespace, jobObj.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *jobResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	name := inputResource.GetMetadata().GetName()
	namespace := inputResource.GetMetadata().GetNamespace()
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	data, err := shared.GetJsonPatchData(inputResource)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting status json patch data")
	}
	jobObj, err := rc.Kube.BatchV1().Jobs(namespace).Patch(opts.Ctx, name, types.JSONPatchType, data, metav1.PatchOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "patching job status from kubernetes")
	}
	resource := FromKubeJob(jobObj)

	if resource == nil {
		return nil, errors.Errorf("job %v is not kind %v", name, rc.Kind())
	}
	return resource, nil
}

func (rc *jobResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(opts.Ctx, namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr("", name)
		}
		return nil
	}

	if err := rc.Kube.BatchV1().Jobs(namespace).Delete(opts.Ctx, name, metav1.DeleteOptions{}); err != nil {
		return errors.Wrapf(err, "deleting jobObj %v", name)
	}
	return nil
}

func (rc *jobResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	jobObjList, err := rc.cache.JobLister().Jobs(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing jobs level")
	}
	var resourceList resources.ResourceList
	for _, jobObj := range jobObjList {
		resource := FromKubeJob(jobObj)

		if resource == nil {
			continue
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *jobResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *jobResourceClient) exist(ctx context.Context, namespace, name string) bool {
	_, err := rc.Kube.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	return err == nil
}
