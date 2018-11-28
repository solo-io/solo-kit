package fsutils

import (
	"os"
	"path/filepath"
)

// HomeDir returns the current users home directory irrespecitve of the OS
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// ConfigDir returns the config directory for glooctl
func ConfigDir() (string, error) {
	d := filepath.Join(HomeDir(), ".glooctl")
	_, err := os.Stat(d)
	if err == nil {
		return d, nil
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(d, 0755)
		if err != nil {
			return "", err
		}
		return d, nil
	}

	return d, err
}
