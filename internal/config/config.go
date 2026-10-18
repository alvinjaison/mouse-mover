package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	MinDurationMinutes   = 1
	MaxDurationMinutes   = 24 * 60 // 1440 = 24 hours
	MinMoveIntervalSecs  = 10
	MaxMoveIntervalSecs  = 300
	DefaultMoveIntervalS = 30
	MinDelta             = 1
	MaxDelta             = 20
)

var validPatterns = map[string]bool{
	"nudge":  true,
	"jiggle": true,
	"square": true,
}

// Config holds persistent user settings.
type Config struct {
	// DurationMinutes is how long the mover stays active before auto-stopping.
	DurationMinutes int `json:"duration_minutes"`
	// MoveIntervalSeconds is the pause between each mouse movement (fixed, not user-facing in presets).
	MoveIntervalSeconds int    `json:"move_interval_seconds"`
	Pattern             string `json:"pattern"`
	DeltaPixels         int    `json:"delta_pixels"`
	StartMinimized      bool   `json:"start_minimized"`
	AutoStart           bool   `json:"auto_start"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		DurationMinutes:     0, // 0 = indefinite (no auto-stop)
		MoveIntervalSeconds: DefaultMoveIntervalS,
		Pattern:             "nudge",
		DeltaPixels:         2,
	}
}

// Validate returns an error describing the first invalid field, or nil.
func (c Config) Validate() error {
	// DurationMinutes == 0 means indefinite (valid). If set, must be in range.
	if c.DurationMinutes < 0 {
		return fmt.Errorf("duration %dm is negative", c.DurationMinutes)
	}
	if c.DurationMinutes > MaxDurationMinutes {
		return fmt.Errorf("duration %dm exceeds maximum %dm", c.DurationMinutes, MaxDurationMinutes)
	}
	if c.MoveIntervalSeconds < MinMoveIntervalSecs {
		return fmt.Errorf("move interval %ds is below minimum %ds", c.MoveIntervalSeconds, MinMoveIntervalSecs)
	}
	if c.MoveIntervalSeconds > MaxMoveIntervalSecs {
		return fmt.Errorf("move interval %ds exceeds maximum %ds", c.MoveIntervalSeconds, MaxMoveIntervalSecs)
	}
	if !validPatterns[c.Pattern] {
		return fmt.Errorf("unknown pattern %q (valid: nudge, jiggle, square)", c.Pattern)
	}
	if c.DeltaPixels < MinDelta || c.DeltaPixels > MaxDelta {
		return fmt.Errorf("delta %d out of range [%d,%d]", c.DeltaPixels, MinDelta, MaxDelta)
	}
	return nil
}

// Duration returns the DurationMinutes as a time.Duration.
func (c Config) Duration() time.Duration {
	return time.Duration(c.DurationMinutes) * time.Minute
}

// MoveInterval returns the MoveIntervalSeconds as a time.Duration.
func (c Config) MoveInterval() time.Duration {
	return time.Duration(c.MoveIntervalSeconds) * time.Second
}

// settingsPath returns the full path to the settings JSON file.
func settingsPath() (string, error) {
	dir, err := settingsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// Load reads settings from disk. Missing or corrupt files yield Default()
// without returning an error so the caller never has to handle startup failure.
func Load() (Config, error) {
	path, err := settingsPath()
	if err != nil {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), nil
	}
	// Backfill defaults for newly added fields (e.g. upgrading from old config).
	if cfg.MoveIntervalSeconds == 0 {
		cfg.MoveIntervalSeconds = DefaultMoveIntervalS
	}
	if cfg.DeltaPixels == 0 {
		cfg.DeltaPixels = 2
	}
	if err := cfg.Validate(); err != nil {
		return Default(), nil
	}
	return cfg, nil
}

// Save validates and writes the config to disk atomically.
func Save(c Config) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("refusing to save invalid config: %w", err)
	}

	path, err := settingsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
