package protodep

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/protodep/api"
	"github.com/solo-io/solo-kit/pkg/protodep/internal/gitcache"
	"github.com/spf13/afero"
)

func NewGitFactory(ctx context.Context) (*gitFactory, error) {
	gitCache, err := gitcache.New(ctx, gitcache.Options{})
	if err != nil {
		return nil, err
	}
	return &gitFactory{
		fs:    afero.NewOsFs(),
		cache: gitCache,
	}, nil
}

type gitFactory struct {
	fs    afero.Fs
	cache gitcache.GitCache
}

func (g *gitFactory) Ensure(ctx context.Context, opts *api.Config) error {
	var imports []*api.GitImport
	for _, v := range opts.GetImports() {
		gitImport := v.GetGit()
		if gitImport != nil {
			imports = append(imports, gitImport)
		}
	}
	return nil
}

func (g *gitFactory) prepareCache(ctx context.Context, imports []*api.GitImport) error {
	return g.cache.Ensure(ctx, imports)
}

func (g *gitFactory) ensure(ctx context.Context, imports []*api.GitImport) error {
	return nil
}
