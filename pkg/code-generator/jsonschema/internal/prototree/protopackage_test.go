package prototree

import (
	"context"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("protopackage", func() {
	var (
		root *ProtoPackage
	)

	BeforeEach(func() {
		root = NewProtoTree(context.TODO())
	})

	Context("creation", func() {
		It("Can create a flat non-recursive tree", func() {
			root.registerType()
		})

		It("Can create a nested recursive tree", func() {

		})
	})
})
