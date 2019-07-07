package webhook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/scheme"
	solov1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/webhook"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"

	apix "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type testConverter struct {
}

func (*testConverter) Convert(src resources.Resource, dst resources.Resource) error {
	_, ok := src.(*v1.MockResource)
	if !ok {
		return errors.New("can't translate src")
	}
	_, ok = dst.(*v2alpha1.MockResource)
	if !ok {
		return errors.New("can't translate src")
	}
	return nil
}

var _ = Describe("Conversion Webhook", func() {

	var (
		kubeWebhook  webhook.KubeWebhook
		respRecorder *httptest.ResponseRecorder
	)

	// var decoder *Decoder

	BeforeEach(func() {
		var err error
		kubeWebhook, err = webhook.NewKubeWebhook(context.TODO(), nil, v2alpha1.MockResourceGVK.GroupKind(), &testConverter{})
		Expect(err).NotTo(HaveOccurred())
		respRecorder = &httptest.ResponseRecorder{
			Body: bytes.NewBuffer(nil),
		}
		Expect(kubeWebhook.InjectScheme(scheme.Scheme)).NotTo(HaveOccurred())
		// var err error
		// decoder, err = NewDecoder(scheme.Scheme)
		// Expect(err).NotTo(HaveOccurred())
	})

	doRequest := func(convReq *apix.ConversionReview) *apix.ConversionReview {
		var payload bytes.Buffer

		Expect(json.NewEncoder(&payload).Encode(convReq)).Should(Succeed())

		convReview := &apix.ConversionReview{}
		req := &http.Request{
			Body: ioutil.NopCloser(bytes.NewReader(payload.Bytes())),
		}
		kubeWebhook.ServeHTTP(respRecorder, req)
		Expect(json.NewDecoder(respRecorder.Result().Body).Decode(convReview)).To(Succeed())
		return convReview
	}

	makeV1Obj := func() *solov1.Resource {
		mockResource := &v1.MockResource{
			Data: "hello",
			Metadata: core.Metadata{
				Name:      "one",
				Namespace: "one",
			},
		}
		return v1.MockResourceCrd.KubeResource(mockResource)
	}

	makeV2Obj := func() *solov1.Resource {
		mockResource := &v2alpha1.MockResource{
			Metadata: core.Metadata{
				Name:      "two",
				Namespace: "two",
			},
		}
		return v2alpha1.MockResourceCrd.KubeResource(mockResource)
	}

	It("should convert spoke to hub successfully", func() {

		v1Obj := makeV1Obj()
		v2obj := makeV2Obj()

		convReq := &apix.ConversionReview{
			TypeMeta: metav1.TypeMeta{},
			Request: &apix.ConversionRequest{
				DesiredAPIVersion: v2alpha1.MockResourceGVK.GroupVersion().String(),
				Objects: []runtime.RawExtension{
					{
						Object: v1Obj,
					},
					{
						Object: v2obj,
					},
				},
			},
		}

		convReview := doRequest(convReq)

		Expect(convReview.Response.ConvertedObjects).To(HaveLen(1))
		Expect(convReview.Response.Result.Status).To(Equal(metav1.StatusSuccess))
		// got, _, err := decoder.Decode(convReview.Response.ConvertedObjects[0].Raw)
		// Expect(err).NotTo(HaveOccurred())
		// Expect(got).To(Equal(expected))
	})

})
