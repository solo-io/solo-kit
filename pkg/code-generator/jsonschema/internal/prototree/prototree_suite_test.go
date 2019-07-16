package prototree

import (
	"path/filepath"
	"testing"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"
)

func TestPrototree(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prototree Suite")
}

var (
	desc     []*parser.DescriptorWithPath
	messages []*descriptor.DescriptorProto

	pathFromRoot  = filepath.Join(parser.GopathSrc(), "github.com/solo-io/solo-kit/test/mocks/api")
	customImports = []string{filepath.Join(parser.GopathSrc(), "github.com/solo-io/solo-kit/api/v1")}

	_ = BeforeSuite(func() {
		var err error
		collector := parser.NewCollector(pathFromRoot)
		desc, err = collector.CollectDescriptors(customImports, nil, nil, func(x string) bool { return true })
		Expect(err).NotTo(HaveOccurred())
		Expect(desc).NotTo(HaveLen(0))
		for _, v := range desc {
			for _, msg := range v.GetMessageType() {
				messages = append(messages, msg)
			}
		}
	})
)
