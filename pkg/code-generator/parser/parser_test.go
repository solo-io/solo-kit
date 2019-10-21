package parser_test

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
)

var _ = Describe("DocsGen", func() {

	var (
		uniqueDescriptor     *model.DescriptorWithPath
		duplicateDescriptor1 *model.DescriptorWithPath
		duplicateDescriptor2 *model.DescriptorWithPath
		descriptors          []*model.DescriptorWithPath
	)

	BeforeEach(func() {
		uniqueDescriptor = &model.DescriptorWithPath{
			FileDescriptorProto: &descriptor.FileDescriptorProto{
				Name:    stringPtr("proto1"),
				Options: &descriptor.FileOptions{},
			},
		}
		duplicateDescriptor1 = &model.DescriptorWithPath{
			FileDescriptorProto: &descriptor.FileDescriptorProto{
				Name: stringPtr("proto2"),
			},
		}
		duplicateDescriptor2 = &model.DescriptorWithPath{
			FileDescriptorProto: &descriptor.FileDescriptorProto{
				Name: stringPtr("proto3"),
			},
			ProtoFilePath: "/path/to/proto/file.proto",
		}
	})

	JustBeforeEach(func() {
		descriptors = []*model.DescriptorWithPath{uniqueDescriptor, duplicateDescriptor1, duplicateDescriptor2}
	})

	It("filters duplicate protos", func() {
		expected := []*model.DescriptorWithPath{uniqueDescriptor, duplicateDescriptor1}
		Expect(descriptors).To(HaveLen(3))
		filtered := parser.FilterDuplicateDescriptors(descriptors)
		Expect(filtered).To(Equal(expected))
	})

	It("updates proto filepaths", func() {
		Expect(duplicateDescriptor1.ProtoFilePath).To(BeEmpty())
		filtered := parser.FilterDuplicateDescriptors(descriptors)
		Expect(duplicateDescriptor1.ProtoFilePath).ToNot(BeEmpty())
		Expect(filtered[1]).To(Equal(duplicateDescriptor1))
		Expect(filtered[1].GetName()).To(Equal("proto2"))
		Expect(filtered[1].ProtoFilePath).To(Equal("/path/to/proto/file.proto"))
	})
})

func stringPtr(str string) *string {
	s := str
	return &s
}
