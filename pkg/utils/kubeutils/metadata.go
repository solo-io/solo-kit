package kubeutils

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubetypes "k8s.io/apimachinery/pkg/types"
)

func FromKubeMeta(meta metav1.ObjectMeta) *core.Metadata {
	return &core.Metadata{
		Name:            meta.Name,
		Namespace:       meta.Namespace,
		ResourceVersion: meta.ResourceVersion,
		Labels:          copyMap(meta.Labels),
		Annotations:     copyMap(meta.Annotations),
		Generation:      meta.Generation,
		OwnerReferences: copyKubernetesOwnerReferences(meta.OwnerReferences),
	}
}

func ToKubeMeta(meta *core.Metadata) metav1.ObjectMeta {
	skMeta := ToKubeMetaMaintainNamespace(meta)
	skMeta.Namespace = clients.DefaultNamespaceIfEmpty(meta.Namespace)
	return skMeta
}

func ToKubeMetaMaintainNamespace(meta *core.Metadata) metav1.ObjectMeta {
	if meta == nil {
		return metav1.ObjectMeta{}
	}
	return metav1.ObjectMeta{
		Name:            meta.Name,
		Namespace:       meta.Namespace,
		ClusterName:     meta.Cluster,
		ResourceVersion: meta.ResourceVersion,
		Labels:          copyMap(meta.Labels),
		Annotations:     copyMap(meta.Annotations),
		Generation:      meta.Generation,
		OwnerReferences: copySoloKitOwnerReferences(meta.GetOwnerReferences()),
	}
}

func copyKubernetesOwnerReferences(references []metav1.OwnerReference) []*core.Metadata_OwnerReference {
	result := make([]*core.Metadata_OwnerReference, 0, len(references))
	for _, ref := range references {
		skRef := &core.Metadata_OwnerReference{
			ApiVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
			Uid:        string(ref.UID),
		}
		if ref.Controller != nil {
			skRef.Controller = &wrappers.BoolValue{
				Value: *ref.Controller,
			}
		}
		if ref.BlockOwnerDeletion != nil {
			skRef.BlockOwnerDeletion = &wrappers.BoolValue{
				Value: *ref.BlockOwnerDeletion,
			}
		}
		result = append(result, skRef)
	}
	return result
}

func copySoloKitOwnerReferences(skReferences []*core.Metadata_OwnerReference) []metav1.OwnerReference {
	result := make([]metav1.OwnerReference, 0, len(skReferences))
	for _, skRef := range skReferences {
		ref := metav1.OwnerReference{
			APIVersion: skRef.GetApiVersion(),
			Kind:       skRef.GetKind(),
			Name:       skRef.GetName(),
			UID:        kubetypes.UID(skRef.GetUid()),
		}
		if skRef.GetBlockOwnerDeletion() != nil {
			boolValue := skRef.GetBlockOwnerDeletion().GetValue()
			ref.BlockOwnerDeletion = &boolValue
		}
		if skRef.GetController() != nil {
			boolValue := skRef.GetController().GetValue()
			ref.Controller = &boolValue
		}
		result = append(result, ref)
	}
	return result
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	res := map[string]string{}
	for k, v := range m {
		res[k] = v
	}
	return res
}
