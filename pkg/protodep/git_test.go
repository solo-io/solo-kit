package protodep

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/protodep/api"
)

var _ = FDescribe("protodep", func() {
	var (
		ctx context.Context
		mgr *gitFactory
	)
	BeforeEach(func() {
		ctx = context.Background()
		var err error
		mgr, err = NewGitFactory(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("vendor protos", func() {
		It("can vendor protos", func() {
			err := mgr.prepareCache(ctx, []*api.GitImport{
				{
					Owner:    "gogo",
					Repo:     "protobuf",
					Revision: &api.GitImport_Tag{Tag: "v1.3.0"},
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
