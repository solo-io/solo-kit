package writer

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
)

var commentPrefixes = map[string]string{
	".go":    "//",
	".proto": "//",
	".js":    "//",
	".ts":    "//",
}

const defaultFilePermission = 0644

type FileWriter interface {
	WriteFiles(files []code_generator.File) error
	WriteFile(file code_generator.File) error
}

// writes to the filesystem
type DefaultFileWriter struct {
	Root   string
	Header string // prepended to files before write
}

func (w *DefaultFileWriter) WriteFiles(files []code_generator.File) error {
	for _, file := range files {
		if err := w.WriteFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (w *DefaultFileWriter) WriteFile(file code_generator.File) error {
	name := filepath.Join(w.Root, file.Filename)
	content := file.Content

	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return err
	}

	log.Printf("Writing %v", name)

	commentPrefix := commentPrefixes[filepath.Ext(name)]
	if commentPrefix == "" {
		// set default comment char to "#" as this is the most common
		commentPrefix = "#"
	}

	if w.Header != "" {
		content = fmt.Sprintf("%s %s\n\n", commentPrefix, w.Header) + content
	}

	if err := ioutil.WriteFile(name, []byte(content), defaultFilePermission); err != nil {
		return err
	}
	return nil
}
