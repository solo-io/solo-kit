// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"fmt"
	"hash"
	"hash/fnv"
	"log"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KubeconfigsSnapshot struct {
	Kubeconfigs KubeConfigList
}

func (s KubeconfigsSnapshot) Clone() KubeconfigsSnapshot {
	return KubeconfigsSnapshot{
		Kubeconfigs: s.Kubeconfigs.Clone(),
	}
}

func (s KubeconfigsSnapshot) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	if _, err := s.hashKubeconfigs(hasher); err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (s KubeconfigsSnapshot) hashKubeconfigs(hasher hash.Hash64) (uint64, error) {
	return hashutils.HashAllSafe(hasher, s.Kubeconfigs.AsInterfaces()...)
}

func (s KubeconfigsSnapshot) HashFields() []zap.Field {
	var fields []zap.Field
	hasher := fnv.New64()
	KubeconfigsHash, err := s.hashKubeconfigs(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("kubeconfigs", KubeconfigsHash))
	snapshotHash, err := s.Hash(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return append(fields, zap.Uint64("snapshotHash", snapshotHash))
}

func (s *KubeconfigsSnapshot) GetResourcesList(resource resources.Resource) (resources.ResourceList, error) {
	switch resource.(type) {
	case *KubeConfig:
		return s.Kubeconfigs.AsResources(), nil
	default:
		return resources.ResourceList{}, eris.New("did not contain the input resource type returning empty list")
	}
}

func (s *KubeconfigsSnapshot) RemoveFromResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key()
	switch resource.(type) {
	case *KubeConfig:

		for i, res := range s.Kubeconfigs {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Kubeconfigs = append(s.Kubeconfigs[:i], s.Kubeconfigs[i+1:]...)
				break
			}
		}
		return nil
	default:
		return eris.Errorf("did not remove the resource because its type does not exist [%T]", resource)
	}
}

func (s *KubeconfigsSnapshot) UpsertToResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key()
	switch typed := resource.(type) {
	case *KubeConfig:
		updated := false
		for i, res := range s.Kubeconfigs {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Kubeconfigs[i] = typed
				updated = true
			}
		}
		if !updated {
			s.Kubeconfigs = append(s.Kubeconfigs, typed)
		}
		s.Kubeconfigs.Sort()
		return nil
	default:
		return eris.Errorf("did not add/replace the resource type because it does not exist %T", resource)
	}
}

type KubeconfigsSnapshotStringer struct {
	Version     uint64
	Kubeconfigs []string
}

func (ss KubeconfigsSnapshotStringer) String() string {
	s := fmt.Sprintf("KubeconfigsSnapshot %v\n", ss.Version)

	s += fmt.Sprintf("  Kubeconfigs %v\n", len(ss.Kubeconfigs))
	for _, name := range ss.Kubeconfigs {
		s += fmt.Sprintf("    %v\n", name)
	}

	return s
}

func (s KubeconfigsSnapshot) Stringer() KubeconfigsSnapshotStringer {
	snapshotHash, err := s.Hash(nil)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return KubeconfigsSnapshotStringer{
		Version:     snapshotHash,
		Kubeconfigs: s.Kubeconfigs.NamespacesDotNames(),
	}
}

var KubeconfigsGvkToHashableResource = map[schema.GroupVersionKind]func() resources.HashableResource{
	KubeConfigGVK: NewKubeConfigHashableResource,
}
