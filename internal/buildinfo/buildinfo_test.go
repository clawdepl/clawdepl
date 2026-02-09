package buildinfo

import (
	"testing"
)

func TestMode(t *testing.T) {
	mode := Mode()
	// In tests without -tags debug, Debug is false
	if Debug {
		if mode != "debug" {
			t.Errorf("Mode() = %q, want %q when Debug=true", mode, "debug")
		}
	} else {
		if mode != "prod" {
			t.Errorf("Mode() = %q, want %q when Debug=false", mode, "prod")
		}
	}
}

func TestIsRelease(t *testing.T) {
	// Save original value
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version string
		want    bool
	}{
		{"dev", false},
		{"1.0.0", true},
		{"0.1.0", true},
		{"2.3.4-beta", true},
	}

	for _, tt := range tests {
		Version = tt.version
		got := IsRelease()
		if got != tt.want {
			t.Errorf("IsRelease() with Version=%q = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestShortCommit(t *testing.T) {
	// Save original value
	origCommit := Commit
	defer func() { Commit = origCommit }()

	tests := []struct {
		commit string
		want   string
	}{
		{"unknown", "unknown"},
		{"abc1234", "abc1234"},
		{"abc12345678901234567890", "abc1234"},
		{"abcdefghijklmnop", "abcdefg"},
		{"short", "short"},
	}

	for _, tt := range tests {
		Commit = tt.commit
		got := ShortCommit()
		if got != tt.want {
			t.Errorf("ShortCommit() with Commit=%q = %q, want %q", tt.commit, got, tt.want)
		}
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that default values are sensible for local development
	if Name == "" {
		t.Error("Name should not be empty")
	}
	if Name != "clawdpl" {
		t.Errorf("Name = %q, want %q", Name, "clawdpl")
	}

	// DefaultEndpoint should be a valid URL
	if DefaultEndpoint == "" {
		t.Error("DefaultEndpoint should not be empty")
	}

	// GoVersion should be set by runtime
	if GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	// GOOS and GOARCH should be set by runtime
	if GOOS == "" {
		t.Error("GOOS should not be empty")
	}
	if GOARCH == "" {
		t.Error("GOARCH should not be empty")
	}
}
