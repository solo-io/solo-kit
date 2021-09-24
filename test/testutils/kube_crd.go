package testutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetHelmCustomResourceDefinition(skCrd crd.Crd, labels map[string]string) *v1beta1.CustomResourceDefinition {
	scope := v1beta1.NamespaceScoped
	if skCrd.ClusterScoped {
		scope = v1beta1.ClusterScoped
	}
	return &v1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: skCrd.FullName(),
			Annotations: map[string]string{
				"helm.sh/hook": "crd-install",
			},
			Labels: labels,
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: skCrd.Group,
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:       skCrd.KindName,
				ListKind:   skCrd.KindName + "List",
				Plural:     skCrd.Plural,
				ShortNames: []string{skCrd.ShortName},
			},
			Scope:   scope,
			Version: "v1",
		},
	}
}
