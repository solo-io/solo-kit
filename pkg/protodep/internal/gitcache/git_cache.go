package gitcache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/solo-kit/pkg/protodep/api"
	"github.com/solo-io/solo-kit/pkg/protodep/internal/codefetcher"
	"github.com/spf13/afero"
)

const (
	DefaultConfigDir = ".protodep"
)

type Options struct {
	WorkingDirectory string
	Fs               afero.Fs
	Client           *github.Client
}

func (o Options) saneDefaults(ctx context.Context) (Options, error) {
	opts := Options{}
	if opts.WorkingDirectory == "" {
		configDir, err := os.UserHomeDir()
		if err != nil {
			return Options{}, err
		}
		opts.WorkingDirectory = filepath.Join(configDir, DefaultConfigDir)
	}
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	if opts.Client == nil {
		opts.Client = githubutils.GetClientWithOrWithoutToken(ctx)
	}
	return opts, nil
}

type GitCache interface {
	Ensure(ctx context.Context, imports []*api.GitImport) error
}

func New(ctx context.Context, opts Options) (cache *gitCache, err error) {
	opts, err = opts.saneDefaults(ctx)
	if err != nil {
		return nil, err
	}
	fetcher := codefetcher.New(opts.Client, opts.Fs)
	return &gitCache{
		fetcher:          fetcher,
		fs:               opts.Fs,
		workingDirectory: opts.WorkingDirectory,
		available:        nil,
	}, nil
}

type cachedRepoId struct {
	owner string
	repo  string
	sha   string
}

type gitCache struct {
	fetcher codefetcher.CodeFetcher
	fs      afero.Fs

	workingDirectory string
	available        map[cachedRepoId]codefetcher.CachedRepo
}

func (g *gitCache) translateCacheIds(imports []*api.GitImport) []cachedRepoId {
	var cachedRepoIds []cachedRepoId
	for _, v := range imports {
		cacheId := cachedRepoId{
			owner: v.GetOwner(),
			repo:  v.GetRepo(),
		}
		switch typedRevision := v.GetRevision().(type) {
		case *api.GitImport_Sha:
			cacheId.sha = typedRevision.Sha
		case *api.GitImport_Tag:
			cacheId.sha = typedRevision.Tag
		}
		cachedRepoIds = append(cachedRepoIds, cacheId)
	}
	return cachedRepoIds
}

func dirFromCachedRepoId(id cachedRepoId, rootPath string) string {
	return filepath.Join(rootPath, "github.com", id.owner, fmt.Sprintf("%s@%s", id.repo, id.sha))
}

func (g *gitCache) ensureCachedRepo(ctx context.Context, id cachedRepoId) (codefetcher.CachedRepo, error) {
	// check if repo already exists in cache, this
	if v, ok := g.available[id]; ok {
		return v, nil
	}
	directory := dirFromCachedRepoId(id, g.workingDirectory)
	cachedRepo, err := g.fetcher.Fetch(ctx, directory, id.owner, id.repo, id.sha)
	if err != nil {
		return nil, err
	}
	g.available[id] = cachedRepo
	return cachedRepo, nil
}

func (g *gitCache) reconcileLocalCache() error {
	afero.Walk(g.fs, g.workingDirectory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		return nil
	})
	return nil
}

func (g *gitCache) Ensure(ctx context.Context, imports []*api.GitImport) error {
	cachedRepoIds := g.translateCacheIds(imports)
	// ensure cache directory exists
	if err := g.fs.MkdirAll(g.workingDirectory, 0777); err != nil {
		return err
	}
	if err := g.reconcileLocalCache(); err != nil {
		return err
	}
	var cachedRepos []codefetcher.CachedRepo
	for _, v := range cachedRepoIds {
		cachedRepo, err := g.ensureCachedRepo(ctx, v)
		if err != nil {
			return err
		}
		cachedRepos = append(cachedRepos, cachedRepo)
	}
	return nil
}
