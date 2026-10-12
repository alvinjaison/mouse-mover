//go:build !windows && !darwin

package config

import (
	"os"
	"path/filepath"
)

func settingsDir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "MouseMover"), nil
}
