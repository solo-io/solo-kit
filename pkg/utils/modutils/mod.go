package modutils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

func GetCurrentModPackageFile() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	modBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(modBytes), nil
}
