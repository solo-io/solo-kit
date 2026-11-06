package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type recordedRequest struct {
	namespace string
	selector  string
}

// fieldSelectorRecorder records the namespace and field selector of every list
// and watch request the cache sends, keyed by resource.
type fieldSelectorRecorder struct {
	mu       sync.Mutex
	requests map[string][]recordedRequest
}

func (r *fieldSelectorRecorder) record(action k8stesting.Action, selector fields.Selector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value := ""
	if selector != nil {
		value = selector.String()
	}
	resource := action.GetResource().Resource
	r.requests[resource] = append(r.requests[resource], recordedRequest{namespace: action.GetNamespace(), selector: value})
}

func (r *fieldSelectorRecorder) getRequests(resource string) []recordedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedRequest(nil), r.requests[resource]...)
}

func (r *fieldSelectorRecorder) get(resource string) []string {
	var selectors []string
	for _, req := range r.getRequests(resource) {
		selectors = append(selectors, req.selector)
	}
	return selectors
}

func newRecordingClient(r *fieldSelectorRecorder) *fake.Clientset {
	client := fake.NewClientset()
	client.PrependReactor("list", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		r.record(action, action.(k8stesting.ListAction).GetListRestrictions().Fields)
		return false, nil, nil
	})
	client.PrependWatchReactor("*", func(action k8stesting.Action) (bool, watch.Interface, error) {
		r.record(action, action.(k8stesting.WatchAction).GetWatchRestrictions().Fields)
		// The fake cannot serve a watch-list stream, so fail it and let the
		// reflector fall back to LIST and WATCH, which are recorded too.
		if impl, ok := action.(k8stesting.WatchActionImpl); ok && impl.ListOptions.SendInitialEvents != nil && *impl.ListOptions.SendInitialEvents {
			return true, nil, errors.New("watch-list is not supported by the fake clientset")
		}
		return false, nil, nil
	})
	return client
}

func TestWithSecretsFieldSelector(t *testing.T) {
	for _, tc := range []struct {
		name     string
		opts     []Option
		selector string
	}{
		{"no option", nil, ""},
		{"nil selector", []Option{WithSecretsFieldSelector(nil)}, ""},
		{"empty selector", []Option{WithSecretsFieldSelector(fields.Everything())}, ""},
		{"single term", []Option{WithSecretsFieldSelector(
			fields.OneTermNotEqualSelector("type", "helm.sh/release.v1"),
		)}, "type!=helm.sh/release.v1"},
		{"multiple terms", []Option{WithSecretsFieldSelector(fields.AndSelectors(
			fields.OneTermNotEqualSelector("type", "helm.sh/release.v1"),
			fields.OneTermNotEqualSelector("type", "kubernetes.io/service-account-token"),
		))}, "type!=helm.sh/release.v1,type!=kubernetes.io/service-account-token"},
		{"last option wins", []Option{
			WithSecretsFieldSelector(fields.OneTermNotEqualSelector("type", "helm.sh/release.v1")),
			WithSecretsFieldSelector(nil),
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &fieldSelectorRecorder{requests: map[string][]recordedRequest{}}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if _, err := NewKubeCoreCacheWithOptions(ctx, newRecordingClient(recorder), time.Hour, []string{metav1.NamespaceAll}, tc.opts...); err != nil {
				t.Fatalf("creating cache: %v", err)
			}

			secretRequests := recorder.get("secrets")
			if len(secretRequests) == 0 {
				t.Fatal("cache sent no list or watch request for secrets")
			}
			for _, got := range secretRequests {
				if got != tc.selector {
					t.Errorf("secrets request field selector = %q, want %q", got, tc.selector)
				}
			}
			// The selector applies to Secrets only.
			for _, resource := range []string{"pods", "services", "configmaps", "namespaces"} {
				for _, got := range recorder.get(resource) {
					if got != "" {
						t.Errorf("%s request field selector = %q, want none", resource, got)
					}
				}
			}
		})
	}
}

func TestWithSecretsFieldSelectorPerNamespace(t *testing.T) {
	recorder := &fieldSelectorRecorder{requests: map[string][]recordedRequest{}}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	selector := fields.OneTermNotEqualSelector("type", "helm.sh/release.v1")
	if _, err := NewKubeCoreCacheWithOptions(ctx, newRecordingClient(recorder), time.Hour, []string{"ns1", "ns2"}, WithSecretsFieldSelector(selector)); err != nil {
		t.Fatalf("creating cache: %v", err)
	}

	namespaces := map[string]bool{}
	for _, req := range recorder.getRequests("secrets") {
		namespaces[req.namespace] = true
		if req.selector != selector.String() {
			t.Errorf("secrets request in %q field selector = %q, want %q", req.namespace, req.selector, selector.String())
		}
	}
	if !namespaces["ns1"] || !namespaces["ns2"] {
		t.Errorf("secrets requests covered namespaces %v, want ns1 and ns2", namespaces)
	}
}
