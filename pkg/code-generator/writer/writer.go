package writer

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	codegenerator "github.com/solo-io/solo-kit/pkg/code-generator"
)

const DefaultFileHeader = `Code generated by solo-kit. DO NOT EDIT.`
const NoFileHeader = ""

type FileWriter interface {
	WriteFile(file codegenerator.File) error
	WriteFiles(files codegenerator.Files) error
}

// writes to the filesystem
type DefaultFileWriter struct {
	Root               string
	HeaderFromFilename func(string) string // prepended to files before write
}

func (w *DefaultFileWriter) WriteFiles(files codegenerator.Files) error {
	for _, file := range files {
		if err := w.WriteFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (w *DefaultFileWriter) WriteFile(file codegenerator.File) error {
	name := filepath.Join(w.Root, file.Filename)
	content := file.Content

	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return err
	}

	perm := file.Permission
	if perm == 0 {
		perm = 0644
	}

	log.Printf("Writing %v", name)

	if w.HeaderFromFilename != nil {
		header := w.HeaderFromFilename(file.Filename)
		if header != NoFileHeader {
			content = header + content
		}
	}

	return ioutil.WriteFile(name, []byte(content), perm)
}
