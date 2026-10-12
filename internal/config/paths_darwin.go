//go:build darwin

package config

import (
	"os"
	"path/filepath"
)

func settingsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "MouseMover"), nil
}
