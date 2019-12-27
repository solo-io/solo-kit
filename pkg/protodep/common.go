package protodep

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"
	"github.com/spf13/afero"
)

type FileCopier interface {
	Copy(src, dst string) (int64, error)
}

type copier struct {
	fs afero.Fs
}

func NewCopier(fs afero.Fs) *copier {
	return &copier{
		fs: fs,
	}
}

var (
	IrregularFileError = func(file string) error {
		return eris.Errorf("%s is not a regular file", file)
	}
)

func NewDefaultCopier() *copier {
	return &copier{fs: afero.NewOsFs()}
}

func (c *copier) Copy(src, dst string) (int64, error) {
	log.Printf("copying %v -> %v", src, dst)

	if err := c.fs.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return 0, err
	}

	srcStat, err := c.fs.Stat(src)
	if err != nil {
		return 0, err
	}

	if !srcStat.Mode().IsRegular() {
		return 0, IrregularFileError(src)
	}

	srcFile, err := c.fs.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := c.fs.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}
