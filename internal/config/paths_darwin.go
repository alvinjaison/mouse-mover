//go:build darwin

package config

import (
	"os"
	"path/filepath"
)

func settingsDir() (string, error) {
	// Honor XDG_CONFIG_HOME as an override (used to isolate tests); otherwise
	// use the standard macOS per-user Application Support location.
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "MouseMover"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "MouseMover"), nil
}
