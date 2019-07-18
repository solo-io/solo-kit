package prototree

import (
	"context"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("protopackage", func() {
	var (
		soloiov1    = "core.solo.io.v1"
		googleProto = "google.protobuf"

		root *ProtoPackage
	)

	BeforeEach(func() {
		root = NewProtoTree(context.TODO())
	})

	Context("creation", func() {
		It("Can create a flat non-recursive tree", func() {
			tests := []string{soloiov1 + ".Metadata", soloiov1 + ".Status"}
			for _, v := range tests {
				t, found := messages[v]
				Expect(found).To(BeTrue())
				root.registerType(proto.String(soloiov1), t)
			}
			for _, v := range tests {
				_, found := root.LookupType(v)
				Expect(found).To(BeTrue())
			}
		})

		It("Can create a nested recursive tree", func() {
			tests := []string{
				googleProto + ".DescriptorProto.ExtensionRange",
				googleProto + ".DescriptorProto.ReservedRange",
				googleProto + ".DescriptorProto",
			}
			t, found := messages[tests[len(tests)-1]]
			Expect(found).To(BeTrue())
			root.RegisterMessage(proto.String(googleProto), t)
			for _, v := range tests {
				_, found := root.LookupType(v)
				Expect(found).To(BeTrue())
			}

		})
	})
})
