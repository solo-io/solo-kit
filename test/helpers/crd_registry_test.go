package helpers

import (
	"context"
	"testing"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	resourcev1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubetesting "k8s.io/client-go/testing"
)

func TestRegisterCrdRetriesWhilePreviousDefinitionIsTerminating(t *testing.T) {
	testCrd := crd.Crd{
		CrdMeta: crd.CrdMeta{
			Plural:    "widgets",
			Group:     "registration.test",
			KindName:  "Widget",
			ShortName: "wd",
		},
		Version: crd.Version{
			Version: "v1",
			Type:    &resourcev1.Resource{},
		},
	}

	registry := &crdRegistry{}
	if err := registry.addCrd(testCrd); err != nil {
		t.Fatalf("add CRD to registry: %v", err)
	}

	deletionTime := metav1.Now()
	terminating := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              testCrd.FullName(),
			DeletionTimestamp: &deletionTime,
		},
		Status: apiextv1.CustomResourceDefinitionStatus{
			Conditions: []apiextv1.CustomResourceDefinitionCondition{{
				Type:   apiextv1.Established,
				Status: apiextv1.ConditionTrue,
			}},
		},
	}

	clientset := apiextfake.NewSimpleClientset()
	createAttempts := 0
	var registered *apiextv1.CustomResourceDefinition
	clientset.Fake.PrependReactor("create", "customresourcedefinitions", func(action kubetesting.Action) (bool, runtime.Object, error) {
		createAttempts++
		if createAttempts == 1 {
			return true, nil, apierrors.NewAlreadyExists(schema.GroupResource{
				Group:    apiextv1.GroupName,
				Resource: "customresourcedefinitions",
			}, testCrd.FullName())
		}

		registered = action.(kubetesting.CreateAction).GetObject().(*apiextv1.CustomResourceDefinition).DeepCopy()
		registered.Status.Conditions = []apiextv1.CustomResourceDefinitionCondition{{
			Type:   apiextv1.Established,
			Status: apiextv1.ConditionTrue,
		}}
		return true, registered, nil
	})
	clientset.Fake.PrependReactor("get", "customresourcedefinitions", func(kubetesting.Action) (bool, runtime.Object, error) {
		if registered == nil {
			return true, terminating.DeepCopy(), nil
		}
		return true, registered.DeepCopy(), nil
	})

	if err := registry.registerCrd(context.Background(), testCrd.GroupVersionKind(), clientset); err != nil {
		t.Fatalf("register CRD: %v", err)
	}
	if createAttempts != 2 {
		t.Fatalf("expected registration to retry once, got %d create attempts", createAttempts)
	}
}
