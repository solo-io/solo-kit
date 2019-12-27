package codefetcher

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/solo-io/go-utils/vfsutils"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type CachedRepo interface {
	vfsutils.MountedRepo
	Location() string
}

type explicitMountedRepo struct {
	owner string
	repo  string
	sha   string

	fs           afero.Fs
	repoRootPath string
	client       *github.Client
}

func NewExplicitMountedRepo(ctx context.Context, owner string, repo string, sha string, fs afero.Fs,
	repoRootPath string, client *github.Client) (*explicitMountedRepo, error) {
	emr := &explicitMountedRepo{owner: owner, repo: repo, sha: sha, fs: fs, repoRootPath: repoRootPath, client: client}
	if err := emr.ensureCodeMounted(ctx); err != nil {
		return nil, err
	}
	return emr, nil
}

func (e *explicitMountedRepo) Location() string {
	return e.repoRootPath
}

func (e *explicitMountedRepo) GetOwner() string {
	return e.owner
}

func (e *explicitMountedRepo) GetRepo() string {
	return e.repo
}

func (e *explicitMountedRepo) GetSha() string {
	return e.sha
}

func (e *explicitMountedRepo) GetFileContents(ctx context.Context, path string) ([]byte, error) {
	if err := e.ensureCodeMounted(ctx); err != nil {
		return nil, err
	}
	return afero.ReadFile(e.fs, filepath.Join(e.repoRootPath, path))
}

func (e *explicitMountedRepo) ListFiles(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := e.ensureCodeMounted(ctx); err != nil {
		return nil, err
	}
	return afero.ReadDir(e.fs, filepath.Join(e.repoRootPath, path))
}

func (e *explicitMountedRepo) ensureCodeMounted(ctx context.Context) error {
	if e.client == nil {
		return vfsutils.InvalidDefinitionError("must provide a github client if not using a local filesystem")
	}
	contextutils.LoggerFrom(ctx).Infow("downloading repo archive",
		zap.String("owner", e.owner),
		zap.String("repo", e.repo),
		zap.String("sha", e.sha))
	if err := e.mountCodeWithDirectory(ctx); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error mounting github code",
			zap.Error(err),
			zap.String("owner", e.owner),
			zap.String("repo", e.repo),
			zap.String("sha", e.sha))
		return vfsutils.CodeMountingError(err)
	}
	contextutils.LoggerFrom(ctx).Infow("successfully mounted repo archive",
		zap.String("owner", e.owner),
		zap.String("repo", e.repo),
		zap.String("sha", e.sha),
		zap.String("repoRootPath", e.repoRootPath))
	return nil
}

func (e *explicitMountedRepo) mountCodeWithDirectory(ctx context.Context) (err error) {
	tarFile, err := afero.TempFile(e.fs, "", "tar-file-")
	if err != nil {
		return err
	}
	defer e.fs.Remove(tarFile.Name())
	if err := githubutils.DownloadRepoArchive(ctx, e.client, tarFile, e.owner, e.repo, e.sha); err != nil {
		return err
	}
	// attempt to make dir just in case
	if err := e.fs.MkdirAll(e.repoRootPath, 0777); err != nil {
		return err
	}
	return tarutils.Untar(e.repoRootPath, tarFile.Name(), e.fs)
}
