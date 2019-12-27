package codefetcher

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/rotisserie/eris"
	"github.com/spf13/afero"
)

var (
	_ CodeFetcher = new(codeFetcher)

	UnableToGetArchiveLinkError = func(err error) error {
		return eris.Wrapf(err, "Unable to get archive link")
	}
)

type CodeFetcher interface {
	Fetch(ctx context.Context, dir, owner, repo, ref string) (CachedRepo, error)
	GetArchiveUrl(ctx context.Context, owner, repo, ref string) (string, error)
}

func New(client *github.Client, fs afero.Fs) *codeFetcher {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return &codeFetcher{
		client: client,
		fs:     fs,
	}
}

type codeFetcher struct {
	client *github.Client
	fs     afero.Fs
}

func (c *codeFetcher) Fetch(ctx context.Context, dir, owner, repo, sha string) (CachedRepo, error) {
	return NewExplicitMountedRepo(ctx, owner, repo, sha, c.fs, dir, c.client)
}

func (c *codeFetcher) GetArchiveUrl(ctx context.Context, owner, repo, ref string) (string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	archiveURL, _, err := c.client.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, opts)
	if err != nil {
		return "", UnableToGetArchiveLinkError(err)
	}
	return archiveURL.String(), nil
}
