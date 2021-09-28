package testutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetHelmCustomResourceDefinition(skCrd crd.Crd, labels map[string]string) *v1.CustomResourceDefinition {
	scope := v1.NamespaceScoped
	if skCrd.ClusterScoped {
		scope = v1.ClusterScoped
	}
	return &v1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: skCrd.FullName(),
			Annotations: map[string]string{
				"helm.sh/hook": "crd-install",
			},
			Labels: labels,
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: skCrd.Group,
			Names: v1.CustomResourceDefinitionNames{
				Kind:       skCrd.KindName,
				ListKind:   skCrd.KindName + "List",
				Plural:     skCrd.Plural,
				ShortNames: []string{skCrd.ShortName},
			},
			Scope: scope,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Storage: true,
					Served:  true,
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}
}
