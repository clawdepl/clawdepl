package buildinfo

import (
	"strings"
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
	if Name != "clawdepl" {
		t.Errorf("Name = %q, want %q", Name, "clawdepl")
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

func TestEndpointDefaults(t *testing.T) {
	// Test that all endpoint defaults are valid URLs
	endpoints := map[string]string{
		"ConvexEndpoint":      ConvexEndpoint,
		"ProvisionerEndpoint": ProvisionerEndpoint,
		"AuthEndpoint":        AuthEndpoint,
	}

	for name, endpoint := range endpoints {
		if endpoint == "" {
			t.Errorf("%s should not be empty", name)
		}
		if !strings.HasPrefix(endpoint, "https://") {
			t.Errorf("%s = %q, should start with https://", name, endpoint)
		}
	}
}

func TestConvexEndpoint(t *testing.T) {
	// Verify the Convex endpoint has expected format
	if !strings.Contains(ConvexEndpoint, "convex.site") {
		t.Errorf("ConvexEndpoint = %q, expected to contain 'convex.site'", ConvexEndpoint)
	}
}

func TestProvisionerEndpoint(t *testing.T) {
	// Verify the Provisioner endpoint has expected format
	if !strings.Contains(ProvisionerEndpoint, "railway.app") {
		t.Errorf("ProvisionerEndpoint = %q, expected to contain 'railway.app'", ProvisionerEndpoint)
	}
}

func TestAuthEndpoint(t *testing.T) {
	// Verify the Auth endpoint has expected format
	if !strings.Contains(AuthEndpoint, "vercel.app") {
		t.Errorf("AuthEndpoint = %q, expected to contain 'vercel.app'", AuthEndpoint)
	}
}
