// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"sort"

	github_com_solo_io_solo_kit_api_external_kubernetes_job "github.com/solo-io/solo-kit/api/external/kubernetes/job"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewJob(namespace, name string) *Job {
	job := &Job{}
	job.Job.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return job
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableJob interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_external_kubernetes_job.Job
}

var _ CloneableJob = &github_com_solo_io_solo_kit_api_external_kubernetes_job.Job{}

type Job struct {
	github_com_solo_io_solo_kit_api_external_kubernetes_job.Job
}

func (r *Job) Clone() resources.Resource {
	return &Job{Job: *r.Job.Clone()}
}

func (r *Job) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.Job.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (r *Job) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *Job) GroupVersionKind() schema.GroupVersionKind {
	return JobGVK
}

type JobList []*Job

func (list JobList) Find(namespace, name string) (*Job, error) {
	for _, job := range list {
		if job.GetMetadata().Name == name && job.GetMetadata().Namespace == namespace {
			return job, nil
		}
	}
	return nil, errors.Errorf("list did not find job %v.%v", namespace, name)
}

func (list JobList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, job := range list {
		ress = append(ress, job)
	}
	return ress
}

func (list JobList) Names() []string {
	var names []string
	for _, job := range list {
		names = append(names, job.GetMetadata().Name)
	}
	return names
}

func (list JobList) NamespacesDotNames() []string {
	var names []string
	for _, job := range list {
		names = append(names, job.GetMetadata().Namespace+"."+job.GetMetadata().Name)
	}
	return names
}

func (list JobList) Sort() JobList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list JobList) Clone() JobList {
	var jobList JobList
	for _, job := range list {
		jobList = append(jobList, resources.Clone(job).(*Job))
	}
	return jobList
}

func (list JobList) Each(f func(element *Job)) {
	for _, job := range list {
		f(job)
	}
}

func (list JobList) EachResource(f func(element resources.Resource)) {
	for _, job := range list {
		f(job)
	}
}

func (list JobList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *Job) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	JobGVK = schema.GroupVersionKind{
		Version: "kubernetes",
		Group:   "kubernetes.solo.io",
		Kind:    "Job",
	}
)
