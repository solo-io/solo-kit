package cache

import (
	"context"
	"maps"
	"slices"
	"sort"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// request is a list or watch request the cache sent, with its options as sent.
type request struct {
	verb, resource, namespace, fieldSelector string
	watchList                                bool
}

func requests(client *fake.Clientset) []request {
	var result []request
	for _, action := range client.Actions() {
		switch a := action.(type) {
		case k8stesting.ListActionImpl:
			result = append(result, request{"list", a.GetResource().Resource, a.GetNamespace(), a.ListOptions.FieldSelector, false})
		case k8stesting.WatchActionImpl:
			watchList := a.ListOptions.SendInitialEvents != nil && *a.ListOptions.SendInitialEvents
			result = append(result, request{"watch", a.GetResource().Resource, a.GetNamespace(), a.ListOptions.FieldSelector, watchList})
		}
	}
	return result
}

// watched returns the sorted "resource namespace" pairs that were both listed and
// watched, and fails on watch-list requests: the fake does not support them, so
// the cache must learn that from the client and use LIST and WATCH.
func watched(t *testing.T, reqs []request) []string {
	t.Helper()
	verbs := map[string]map[string]bool{}
	for _, r := range reqs {
		if r.watchList {
			t.Errorf("watch-list request for %s despite a client that does not support it", r.resource)
		}
		key := r.resource + " " + r.namespace
		if verbs[key] == nil {
			verbs[key] = map[string]bool{}
		}
		verbs[key][r.verb] = true
	}
	var result []string
	for key, v := range verbs {
		if !v["list"] || !v["watch"] {
			t.Errorf("%s: list=%t watch=%t, want both", key, v["list"], v["watch"])
		}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func newCache(t *testing.T, client *fake.Clientset, namespaces []string, opts ...Option) *kubeCoreCaches {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	c, err := NewKubeCoreCacheWithOptions(ctx, client, time.Hour, namespaces, opts...)
	if err != nil {
		t.Fatalf("creating cache: %v", err)
	}
	return c
}

func TestDefaultResources(t *testing.T) {
	t.Run("all namespaces", func(t *testing.T) {
		client := fake.NewClientset()
		c := newCache(t, client, []string{metav1.NamespaceAll})
		want := []string{"configmaps ", "namespaces ", "pods ", "secrets ", "services "}
		if got := watched(t, requests(client)); !slices.Equal(got, want) {
			t.Errorf("watched %q, want %q", got, want)
		}
		if c.PodLister() == nil || c.ServiceLister() == nil || c.ConfigMapLister() == nil || c.SecretLister() == nil || c.NamespaceLister() == nil {
			t.Error("a lister is nil")
		}
	})
	t.Run("specific namespaces", func(t *testing.T) {
		client := fake.NewClientset()
		c := newCache(t, client, []string{"ns1", "ns2"})
		var want []string
		for _, resource := range []string{"configmaps", "pods", "secrets", "services"} {
			want = append(want, resource+" ns1", resource+" ns2")
		}
		if got := watched(t, requests(client)); !slices.Equal(got, want) {
			t.Errorf("watched %q, want %q", got, want)
		}
		if c.NamespaceLister() != nil {
			t.Error("NamespaceLister is not nil without all namespaces")
		}
	})
}

func TestWithResources(t *testing.T) {
	t.Run("namespaces only", func(t *testing.T) {
		client := fake.NewClientset(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "existing"}})
		c := newCache(t, client, []string{metav1.NamespaceAll}, WithResources(Namespaces))
		if got, want := watched(t, requests(client)), []string{"namespaces "}; !slices.Equal(got, want) {
			t.Errorf("watched %q, want %q", got, want)
		}
		namespaces, err := c.NamespaceLister().List(labels.Everything())
		if err != nil || len(namespaces) != 1 || namespaces[0].Name != "existing" {
			t.Errorf("NamespaceLister().List() = %v, %v; want the existing namespace", namespaces, err)
		}
		if c.PodLister() != nil || c.ServiceLister() != nil || c.ConfigMapLister() != nil || c.SecretLister() != nil {
			t.Error("a lister of an unwatched resource is not nil")
		}
		if c.NamespacedPodLister("ns") != nil || c.NamespacedServiceLister("ns") != nil || c.NamespacedConfigMapLister("ns") != nil || c.NamespacedSecretLister("ns") != nil {
			t.Error("a namespaced lister of an unwatched resource is not nil")
		}
	})
	t.Run("secrets in specific namespaces", func(t *testing.T) {
		client := fake.NewClientset()
		c := newCache(t, client, []string{"ns1", "ns2"}, WithResources(Secrets))
		if got, want := watched(t, requests(client)), []string{"secrets ns1", "secrets ns2"}; !slices.Equal(got, want) {
			t.Errorf("watched %q, want %q", got, want)
		}
		if c.NamespacedSecretLister("ns1") == nil || c.NamespacedPodLister("ns1") != nil {
			t.Error("want only the Secret listers")
		}
	})
	t.Run("last option wins", func(t *testing.T) {
		client := fake.NewClientset()
		newCache(t, client, []string{metav1.NamespaceAll}, WithResources(Pods), WithResources(Services))
		if got, want := watched(t, requests(client)), []string{"services "}; !slices.Equal(got, want) {
			t.Errorf("watched %q, want %q", got, want)
		}
	})
}

func TestWithFieldSelector(t *testing.T) {
	// The terms are deliberately out of order: the cache must send the selector as
	// given, not re-parsed, which would sort them.
	twoTerms := fields.AndSelectors(
		fields.OneTermNotEqualSelector("type", "kubernetes.io/service-account-token"),
		fields.OneTermNotEqualSelector("type", "helm.sh/release.v1"),
	)
	for _, tc := range []struct {
		name string
		opts []Option
		want map[string]string // resource -> selector; others must have none
	}{
		{"none", nil, nil},
		{"secrets", []Option{WithSecretsFieldSelector(twoTerms)}, map[string]string{"secrets": twoTerms.String()}},
		{"generic", []Option{WithFieldSelector(Pods, fields.OneTermEqualSelector("spec.nodeName", "node-a"))}, map[string]string{"pods": "spec.nodeName=node-a"}},
		{"nil removes", []Option{WithSecretsFieldSelector(twoTerms), WithSecretsFieldSelector(nil)}, nil},
		{"empty removes", []Option{WithSecretsFieldSelector(twoTerms), WithFieldSelector(Secrets, fields.Everything())}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewClientset()
			newCache(t, client, []string{"ns1", "ns2"}, tc.opts...)
			reqs := requests(client)
			watched(t, reqs)
			for _, r := range reqs {
				if want := tc.want[r.resource]; r.fieldSelector != want {
					t.Errorf("%s %s in %q: field selector %q, want %q", r.verb, r.resource, r.namespace, r.fieldSelector, want)
				}
			}
		})
	}
}

func TestWithTransformFunc(t *testing.T) {
	annotated := metav1.ObjectMeta{Name: "a", Namespace: "ns", Annotations: map[string]string{"drop": "x", "keep": "y"}}
	client := fake.NewClientset(&v1.Secret{ObjectMeta: annotated}, &v1.Pod{ObjectMeta: annotated})
	dropAnnotation := func(obj interface{}) (interface{}, error) {
		if accessor, ok := obj.(metav1.ObjectMetaAccessor); ok {
			delete(accessor.GetObjectMeta().GetAnnotations(), "drop")
		}
		return obj, nil
	}
	c := newCache(t, client, []string{metav1.NamespaceAll}, WithTransformFunc(Secrets, dropAnnotation))

	secret, err := c.SecretLister().Secrets("ns").Get("a")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if want := map[string]string{"keep": "y"}; !maps.Equal(secret.Annotations, want) {
		t.Errorf("secret annotations = %v, want %v", secret.Annotations, want)
	}
	pods, err := c.NamespacedPodLister("ns").List(labels.Everything())
	if err != nil || len(pods) != 1 {
		t.Fatalf("list pods = %v, %v", pods, err)
	}
	if want := annotated.Annotations; !maps.Equal(pods[0].Annotations, want) {
		t.Errorf("pod annotations = %v, want %v: the transform applies to Secrets only", pods[0].Annotations, want)
	}
}

func TestOptionValidation(t *testing.T) {
	noop := func(obj interface{}) (interface{}, error) { return obj, nil }
	for _, tc := range []struct {
		name       string
		namespaces []string
		opts       []Option
	}{
		{"no resources", nil, []Option{WithResources()}},
		{"unknown resource", nil, []Option{WithResources("Deployments")}},
		{"selector for unknown resource", nil, []Option{WithFieldSelector("pods", fields.OneTermEqualSelector("spec.nodeName", "a"))}},
		{"nil selector for unknown resource", nil, []Option{WithFieldSelector("pods", nil)}},
		{"empty selector for unknown resource", nil, []Option{WithFieldSelector("pods", fields.Everything())}},
		{"transform for unknown resource", nil, []Option{WithTransformFunc("Secret", noop)}},
		{"nil transform for unknown resource", nil, []Option{WithTransformFunc("Secret", nil)}},
		{"namespaces without all namespaces", []string{"ns1"}, []Option{WithResources(Namespaces)}},
		{"selector for unwatched resource", nil, []Option{WithResources(Secrets), WithFieldSelector(Pods, fields.OneTermEqualSelector("spec.nodeName", "a"))}},
		{"transform for unwatched resource", []string{"ns1"}, []Option{WithTransformFunc(Namespaces, noop)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewClientset()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if _, err := NewKubeCoreCacheWithOptions(ctx, client, time.Hour, tc.namespaces, tc.opts...); err == nil {
				t.Fatal("expected an error")
			}
			if reqs := requests(client); len(reqs) != 0 {
				t.Errorf("sent %d requests despite the invalid options", len(reqs))
			}
		})
	}
}

// A cache watching only Namespaces still notifies subscribers of changes after it
// syncs, which is all the namespace client needs.
func TestWatchEventsNotifySubscribers(t *testing.T) {
	client := fake.NewClientset()
	// The fake ignores resource versions, so it delivers only objects created after
	// the watch is registered. Signal when that has happened.
	watching := make(chan struct{})
	client.PrependWatchReactor("namespaces", func(action k8stesting.Action) (bool, watch.Interface, error) {
		w, err := client.Tracker().Watch(action.GetResource(), action.GetNamespace())
		close(watching)
		return true, w, err
	})
	c := newCache(t, client, []string{metav1.NamespaceAll}, WithResources(Namespaces))
	updates := c.Subscribe()
	defer c.Unsubscribe(updates)

	select {
	case <-watching:
	case <-time.After(10 * time.Second):
		t.Fatal("the cache never watched namespaces")
	}
	if _, err := client.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "added"}}, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-updates:
	case <-time.After(10 * time.Second):
		t.Fatal("no update after creating a namespace")
	}
	if _, err := c.NamespaceLister().Get("added"); err != nil {
		t.Errorf("the lister does not have the added namespace: %v", err)
	}
}
