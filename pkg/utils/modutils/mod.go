package modutils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

/*
	Returns the current go mod package name from the go.mod file.
	Use the function below to get the filename
	Ex: github.com/solo-io/solo-kit
*/
func GetCurrentModPackageName(module string) (string, error) {
	f, err := os.Open(module)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		return "", fmt.Errorf("invalid module file")
	}
	line := scanner.Text()
	parts := strings.Split(line, " ")

	modPath := parts[len(parts)-1]
	if modPath == "/dev/null" || modPath == "" {
		return "", errors.New("solo-kit must be run from within go.mod repo")
	}

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
		return "", err
	}
	return string(modBytes), nil
}
