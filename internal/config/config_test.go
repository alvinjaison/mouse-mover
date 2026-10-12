package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
}

func TestValidateDurationBounds(t *testing.T) {
	tests := []struct {
		name    string
		minutes int
		wantErr bool
	}{
		{"zero (indefinite)", 0, false},
		{"1 min", 1, false},
		{"30m", 30, false},
		{"60m", 60, false},
		{"120m", 120, false},
		{"240m", 240, false},
		{"480m", 480, false},
		{"at max", MaxDurationMinutes, false},
		{"above max", MaxDurationMinutes + 1, true},
		{"negative", -1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.DurationMinutes = tc.minutes
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestValidateMoveInterval(t *testing.T) {
	tests := []struct {
		name    string
		secs    int
		wantErr bool
	}{
		{"below min", 5, true},
		{"at min", MinMoveIntervalSecs, false},
		{"default", DefaultMoveIntervalS, false},
		{"at max", MaxMoveIntervalSecs, false},
		{"above max", MaxMoveIntervalSecs + 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.MoveIntervalSeconds = tc.secs
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"nudge", "nudge", false},
		{"jiggle", "jiggle", false},
		{"square", "square", false},
		{"empty", "", true},
		{"unknown", "zigzag", true},
		{"uppercase", "Nudge", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.Pattern = tc.pattern
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestValidateDeltaBounds(t *testing.T) {
	tests := []struct {
		name    string
		delta   int
		wantErr bool
	}{
		{"below min", 0, true},
		{"at min", MinDelta, false},
		{"at max", MaxDelta, false},
		{"above max", MaxDelta + 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.DeltaPixels = tc.delta
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestPresetDurations(t *testing.T) {
	presets := []int{30, 60, 120, 240, 480}
	for _, p := range presets {
		cfg := Default()
		cfg.DurationMinutes = p
		if err := cfg.Validate(); err != nil {
			t.Errorf("preset %dm should be valid: %v", p, err)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := Config{
		DurationMinutes:     120,
		MoveIntervalSeconds: 30,
		Pattern:             "jiggle",
		DeltaPixels:         5,
		StartMinimized:      true,
		AutoStart:           true,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded != cfg {
		t.Errorf("round-trip mismatch: saved %+v, loaded %+v", cfg, loaded)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("expected default, got %+v", cfg)
	}
}

func TestLoadCorruptReturnsDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "MouseMover")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte("not json"), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("expected default for corrupt file, got %+v", cfg)
	}
}

func TestLoadInvalidValuesReturnsDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, "MouseMover")
	os.MkdirAll(dir, 0o755)
	// Valid JSON but negative duration
	data := `{"duration_minutes": -1, "move_interval_seconds": 30, "pattern": "nudge", "delta_pixels": 2}`
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte(data), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Errorf("expected default for invalid values, got %+v", cfg)
	}
}

func TestSaveRejectsInvalid(t *testing.T) {
	cfg := Default()
	cfg.DurationMinutes = -1
	if err := Save(cfg); err == nil {
		t.Error("expected Save to reject invalid config")
	}
}
