package resources

import (
	"fmt"
	"hash"
	"reflect"
	"sort"

	"github.com/solo-io/protoc-gen-ext/pkg/clone"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

type Resource interface {
	GetMetadata() *core.Metadata
	SetMetadata(meta *core.Metadata)
	Equal(that interface{}) bool
}

type VersionedResource interface {
	Resource
	GroupVersionKind() schema.GroupVersionKind
}

type ProtoResource interface {
	Resource
	proto.Message
}

func ProtoCast(res Resource) (ProtoResource, error) {
	if res == nil {
		return nil, nil
	}
	protoResource, ok := res.(ProtoResource)
	if !ok {
		return nil, errors.Errorf("internal error: unexpected type %T not convertible to resources.Proto", res)
	}
	return protoResource, nil
}

type InputResource interface {
	Resource
	// Deprecated: prefer GetNamespacedStatuses()
	GetStatus() *core.Status
	// Deprecated: prefer SetNamespacedStatuses()
	SetStatus(status *core.Status)
	GetNamespacedStatuses() *core.NamespacedStatuses
	SetNamespacedStatuses(namespacedStatuses *core.NamespacedStatuses)
}

type StatusGetter interface {
	GetStatus(resource InputResource) *core.Status
}

type StatusSetter interface {
	SetStatus(resource InputResource, status *core.Status)
}

type StatusClient interface {
	StatusGetter
	StatusSetter
}

type StatusUnmarshaler interface {
	UnmarshalStatus(status v1.Status, into InputResource)
}

// Custom resources imported in a solo-kit project can implement this interface to control
// how spec and status data is mapped to/from the generic `Resource` type.
type CustomInputResource interface {
	InputResource
	UnmarshalSpec(spec v1.Spec) error
	UnmarshalStatus(status v1.Status, defaultUnmarshaler StatusUnmarshaler)
	MarshalSpec() (v1.Spec, error)
	MarshalStatus() (v1.Status, error)
}

// Hashable is an interface used for hashing the struture.
type Hashable interface {
	Hash(hasher hash.Hash64) (uint64, error)
	MustHash() uint64
}

// HashableResource are Resources that can be hashed
type HashableResource interface {
	Resource
	Hashable
}

type ResourceList []Resource
type ResourcesById map[string]Resource
type ResourcesByKind map[string]ResourceList

func (m ResourcesById) List() ResourceList {
	var all ResourceList
	for _, res := range m {
		all = append(all, res)
	}
	return all.Sort()
}

func (m ResourcesByKind) Add(resources ...Resource) {
	for _, resource := range resources {
		m[Kind(resource)] = append(m[Kind(resource)], resource)
	}
}

func (m ResourcesByKind) Get(resource Resource) []Resource {
	return m[Kind(resource)]
}

func (m ResourcesByKind) List() ResourceList {
	var all ResourceList
	for _, list := range m {
		all = append(all, list...)
	}
	return all.Sort()
}

func (list ResourceList) Contains(list2 ResourceList) bool {
	for _, res2 := range list2 {
		var found bool
		for _, res := range list {
			if res.GetMetadata().Name == res2.GetMetadata().Name && res.GetMetadata().Namespace == res2.GetMetadata().Namespace {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (list ResourceList) Copy() ResourceList {
	var cpy ResourceList
	for _, res := range list {
		cpy = append(cpy, Clone(res))
	}
	return cpy
}

func (list ResourceList) Len() int {
	return len(list)
}

func (list ResourceList) Less(i, j int) bool {
	if result := MetadataCompare(list[i].GetMetadata(), list[j].GetMetadata()); result != 0 {
		return result == -1
	}
	kindi := Kind(list[i])
	kindj := Kind(list[j])
	return kindi < kindj
}

func (list ResourceList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list ResourceList) Sort() ResourceList {
	sorted := make(ResourceList, 0, list.Len())
	for _, res := range list {
		sorted = append(sorted, Clone(res))
	}
	sort.Stable(sorted)
	return sorted
}

func (list ResourceList) Equal(list2 ResourceList) bool {
	if len(list) != len(list2) {
		return false
	}
	for i := range list {
		if !list[i].Equal(list2[i]) {
			return false
		}
	}
	return true
}

func (list ResourceList) FilterByNames(names []string) ResourceList {
	var filtered ResourceList
	for _, resource := range list {
		for _, name := range names {
			if name == resource.GetMetadata().Name {
				filtered = append(filtered, resource)
				break
			}
		}
	}
	return filtered
}

func (list ResourceList) FilterByNamespaces(namespaces []string) ResourceList {
	var filtered ResourceList
	for _, resource := range list {
		for _, namespace := range namespaces {
			if namespace == resource.GetMetadata().Namespace {
				filtered = append(filtered, resource)
				break
			}
		}
	}
	return filtered
}

func (list ResourceList) FilterByKind(kind string) ResourceList {
	var resourcesOfKind ResourceList
	for _, res := range list {
		if Kind(res) == kind {
			resourcesOfKind = append(resourcesOfKind, res)
		}
	}
	return resourcesOfKind
}

func (list ResourceList) FilterByList(list2 ResourceList) ResourceList {
	return list.FilterByNamespaces(list2.Namespaces()).FilterByNames(list.Names())
}

func (list ResourceList) Names() []string {
	var names []string
	for _, resource := range list {
		names = append(names, resource.GetMetadata().Name)
	}
	return names
}

func (list ResourceList) Each(do func(resource Resource)) {
	for i, resource := range list {
		do(resource)
		list[i] = resource
	}
}

func (list ResourceList) EachErr(do func(resource Resource) error) error {
	for i, resource := range list {
		if err := do(resource); err != nil {
			return err
		}
		list[i] = resource
	}
	return nil
}

func (list ResourceList) ByCluster() map[string]ResourceList {
	byCluster := make(map[string]ResourceList)
	list.Each(func(resource Resource) {
		byCluster[resource.GetMetadata().Cluster] = append(
			byCluster[resource.GetMetadata().Cluster], resource)
	})
	return byCluster
}

func (list ResourceList) Find(namespace, name string) (Resource, error) {
	for _, resource := range list {
		if resource.GetMetadata().Name == name {
			if namespace == "" || resource.GetMetadata().Namespace == namespace {
				return resource, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find resource %v.%v", namespace, name)
}
func (list ResourceList) Namespaces() []string {
	var namespaces []string
	for _, resource := range list {
		namespaces = append(namespaces, resource.GetMetadata().Namespace)
	}
	return namespaces
}
func (list ResourceList) AsInputResourceList() InputResourceList {
	var inputs InputResourceList
	for _, res := range list {
		inputRes, ok := res.(InputResource)
		if !ok {
			continue
		}
		inputs = append(inputs, inputRes)
	}
	return inputs
}

func MetadataCompare(metai, metaj *core.Metadata) int {
	if metai.GetCluster() != metaj.GetCluster() {
		if metai.GetCluster() < metaj.GetCluster() {
			return -1
		}
		return 1
	}

	if metai.GetNamespace() != metaj.GetNamespace() {
		if metai.GetNamespace() < metaj.GetNamespace() {
			return -1
		}
		return 1
	}

	if metai.GetName() != metaj.GetName() {
		if metai.GetName() < metaj.GetName() {
			return -1
		}
		return 1
	}
	return 0
}

type InputResourceList []InputResource
type InputResourcesByKind map[string]InputResourceList

func (list InputResourceList) Len() int {
	return len(list)
}

func (list InputResourceList) Less(i, j int) bool {
	if result := MetadataCompare(list[i].GetMetadata(), list[j].GetMetadata()); result != 0 {
		return result == -1
	}

	kindi := Kind(list[i])
	kindj := Kind(list[j])
	return kindi < kindj
}

func (list InputResourceList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (m InputResourcesByKind) Add(resource InputResource) {
	m[Kind(resource)] = append(m[Kind(resource)], resource)
}
func (m InputResourcesByKind) Get(resource InputResource) InputResourceList {
	return m[Kind(resource)]
}
func (m InputResourcesByKind) List() InputResourceList {
	var all InputResourceList
	for _, list := range m {
		all = append(all, list...)
	}
	// sort by type
	sort.Stable(all)
	return all
}
func (list InputResourceList) Contains(list2 InputResourceList) bool {
	for _, res2 := range list2 {
		var found bool
		for _, res := range list {
			if res.GetMetadata().Name == res2.GetMetadata().Name && res.GetMetadata().Namespace == res2.GetMetadata().Namespace {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func (list InputResourceList) Copy() InputResourceList {
	var cpy InputResourceList
	for _, res := range list {
		cpy = append(cpy, Clone(res).(InputResource))
	}
	return cpy
}
func (list InputResourceList) Equal(list2 InputResourceList) bool {
	if len(list) != len(list2) {
		return false
	}
	for i := range list {
		if !list[i].Equal(list2[i]) {
			return false
		}
	}
	return true
}
func (list InputResourceList) FilterByNames(names []string) InputResourceList {
	var filtered InputResourceList
	for _, resource := range list {
		for _, name := range names {
			if name == resource.GetMetadata().Name {
				filtered = append(filtered, resource)
				break
			}
		}
	}
	return filtered
}
func (list InputResourceList) FilterByNamespaces(namespaces []string) InputResourceList {
	var filtered InputResourceList
	for _, resource := range list {
		for _, namespace := range namespaces {
			if namespace == resource.GetMetadata().Namespace {
				filtered = append(filtered, resource)
				break
			}
		}
	}
	return filtered
}
func (list InputResourceList) FilterByKind(kind string) InputResourceList {
	var resourcesOfKind InputResourceList
	for _, res := range list {
		if Kind(res) == kind {
			resourcesOfKind = append(resourcesOfKind, res)
		}
	}
	return resourcesOfKind
}
func (list InputResourceList) FilterByList(list2 InputResourceList) InputResourceList {
	return list.FilterByNamespaces(list2.Namespaces()).FilterByNames(list.Names())
}
func (list InputResourceList) Find(namespace, name string) (InputResource, error) {
	for _, resource := range list {
		if resource.GetMetadata().Name == name {
			if namespace == "" || resource.GetMetadata().Namespace == namespace {
				return resource, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find resource %v.%v", namespace, name)
}
func (list InputResourceList) Names() []string {
	var names []string
	for _, resource := range list {
		names = append(names, resource.GetMetadata().Name)
	}
	return names
}
func (list InputResourceList) Namespaces() []string {
	var namespaces []string
	for _, resource := range list {
		namespaces = append(namespaces, resource.GetMetadata().Namespace)
	}
	return namespaces
}
func (list InputResourceList) AsResourceList() ResourceList {
	var resources ResourceList
	for _, res := range list {
		resources = append(resources, res)
	}
	return resources
}

type CloneableResource interface {
	Resource
	Clone() Resource
}

func Clone(resource Resource) Resource {
	if cloneable, ok := resource.(clone.Cloner); ok {
		return cloneable.Clone().(Resource)
	}
	if cloneable, ok := resource.(CloneableResource); ok {
		return cloneable.Clone()
	}
	if protoMessage, ok := resource.(ProtoResource); ok {
		return proto.Clone(protoMessage).(Resource)
	}
	panic(fmt.Errorf("resource %T is not any of [clone.Cloner, CloneableResource, ProtoResource]", resource))
}

func Kind(resource Resource) string {
	return reflect.TypeOf(resource).String()
}

func UpdateMetadata(resource Resource, updateFunc func(meta *core.Metadata)) {
	meta := resource.GetMetadata()
	updateFunc(meta)
	resource.SetMetadata(meta)
}

func UpdateListMetadata(resources ResourceList, updateFunc func(meta *core.Metadata)) {
	for i, resource := range resources {
		meta := resource.GetMetadata()
		updateFunc(meta)
		resource.SetMetadata(meta)
		resources[i] = resource
	}
}

func Validate(resource Resource) error {
	return ValidateName(resource.GetMetadata().Name)
}

func ValidateName(name string) error {
	errs := validation.IsDNS1123Subdomain(name)
	if len(name) < 1 {
		errs = append(errs, "name cannot be empty. Given: "+name)
	}
	if len(name) > 253 {
		errs = append(errs, "name has a max length of 253 characters. Given: "+name)
	}
	if len(errs) > 0 {
		return errors.Errors(errs)
	}
	return nil
}
