package modutils

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/rotisserie/eris"
)

var (
	ModPackageFileError  = eris.New("could not find mod package file")
	NonGoModPackageError = eris.New("path could not be found, as this function must be run from within a" +
		"go.mod package")

	EmptyFileError = eris.New("empty file supplied, must be")

	UnableToListPackagesError = eris.New("unable to list dependencies for current go.mod packages")
)

/*
	Returns the current go mod package name from the go.mod file.
	Use the function below to get the filename
	Ex: github.com/solo-io/solo-kit
*/
func GetCurrentModPackageName(module string) (string, error) {
	f, err := os.Open(module)
	if err != nil {
		return "", eris.Wrap(ModPackageFileError, err.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		return "", EmptyFileError
	}
	line := scanner.Text()
	parts := strings.Split(line, " ")

	return parts[len(parts)-1], nil
}

/*
	Returns the current go mod package
	Ex: /path/to/solo-kit/go.mod

	Will return /dev/null on unix if not in a go.mod package
*/
func GetCurrentModPackageFile() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	modBytes, err := cmd.Output()
	if err != nil {
		return "", eris.Wrap(ModPackageFileError, err.Error())
	}
	trimmedModFile := strings.TrimSpace(string(modBytes))
	if trimmedModFile == "/dev/null" || trimmedModFile == "" {
		return "", NonGoModPackageError
	}
	return trimmedModFile, nil
}

func GetCurrentPackageList() (*bytes.Buffer, error) {
	modPackageReader := &bytes.Buffer{}
	packageListCmd := exec.Command("go", "list", "-m", "all")
	packageListCmd.Stdout = modPackageReader
	packageListCmd.Stderr = modPackageReader
	err := packageListCmd.Run()
	if err != nil {
		return nil, eris.Wrapf(UnableToListPackagesError, "filename: %s", modPackageReader.String())
	}
	return modPackageReader, nil
}
