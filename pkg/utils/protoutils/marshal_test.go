package protoutils_test

import (
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
)

type testType struct {
	A string
	B b
}

type b struct {
	C string
	D string
}

var tests = []struct {
	in       interface{}
	expected proto.Message
}{
	{
		in: testType{
			A: "a",
			B: b{
				C: "c",
				D: "d",
			},
		},
		expected: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"A": {
					Kind: &structpb.Value_StringValue{StringValue: "a"},
				},
				"B": {
					Kind: &structpb.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"C": {
									Kind: &structpb.Value_StringValue{StringValue: "c"},
								},
								"D": {
									Kind: &structpb.Value_StringValue{StringValue: "d"},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		in: map[string]interface{}{
			"a": "b",
			"c": "d",
		},
		expected: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"a": {
					Kind: &structpb.Value_StringValue{StringValue: "b"},
				},
				"c": {
					Kind: &structpb.Value_StringValue{StringValue: "d"},
				},
			},
		},
	},
}

var _ = Describe("Protoutil Funcs", func() {
	Describe("MarshalStruct", func() {
		//TODO: does not compile, needs fixing
		//for _, test := range tests {
		//	It("returns a pb struct for object of the given type", func() {
		//		pb, err := MarshalStruct(test.in)
		//		Expect(err).NotTo(HaveOccurred())
		//		Expect(pb).To(Equal(test.expected))
		//	})
		//}
	})
})
