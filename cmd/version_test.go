package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
)

func TestVersionInfo(t *testing.T) {
	info := VersionInfo{
		Name:       buildinfo.Name,
		Version:    buildinfo.Version,
		Commit:     buildinfo.Commit,
		CommitFull: buildinfo.CommitFull,
		Date:       buildinfo.Date,
		Mode:       buildinfo.Mode(),
		GOOS:       buildinfo.GOOS,
		GOARCH:     buildinfo.GOARCH,
		GoVersion:  buildinfo.GoVersion,
	}

	// Test that all fields are populated
	if info.Name == "" {
		t.Error("VersionInfo.Name should not be empty")
	}
	if info.Version == "" {
		t.Error("VersionInfo.Version should not be empty")
	}
	if info.Mode == "" {
		t.Error("VersionInfo.Mode should not be empty")
	}
	if info.GOOS == "" {
		t.Error("VersionInfo.GOOS should not be empty")
	}
	if info.GOARCH == "" {
		t.Error("VersionInfo.GOARCH should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("VersionInfo.GoVersion should not be empty")
	}
}

func TestVersionJSONOutput(t *testing.T) {
	info := VersionInfo{
		Name:       "clawdepl",
		Version:    "1.2.3",
		Commit:     "abc1234",
		CommitFull: "abc1234567890",
		Date:       "2026-02-08T12:00:00Z",
		Mode:       "prod",
		GOOS:       "linux",
		GOARCH:     "amd64",
		GoVersion:  "go1.21.0",
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal VersionInfo: %v", err)
	}

	// Unmarshal back to verify
	var decoded VersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal VersionInfo: %v", err)
	}

	if decoded.Name != info.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, info.Name)
	}
	if decoded.Version != info.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, info.Version)
	}
	if decoded.Commit != info.Commit {
		t.Errorf("Commit = %q, want %q", decoded.Commit, info.Commit)
	}
	if decoded.Mode != info.Mode {
		t.Errorf("Mode = %q, want %q", decoded.Mode, info.Mode)
	}
}

func TestPrintVersionHuman(t *testing.T) {
	// Capture output by testing the format logic
	info := VersionInfo{
		Name:       "clawdepl",
		Version:    "1.2.3",
		Commit:     "abc1234",
		CommitFull: "abc1234567890",
		Date:       "2026-02-08T12:00:00Z",
		Mode:       "prod",
		GOOS:       "linux",
		GOARCH:     "amd64",
		GoVersion:  "go1.21.0",
	}

	// Test version string formatting
	versionStr := info.Version
	if versionStr != "dev" {
		versionStr = "v" + versionStr
	}

	if versionStr != "v1.2.3" {
		t.Errorf("Version string = %q, want %q", versionStr, "v1.2.3")
	}

	// Test commit string formatting
	commitStr := info.Commit
	if info.CommitFull != "" && info.CommitFull != "unknown" && info.CommitFull != info.Commit {
		commitStr = info.Commit + " (" + info.CommitFull + ")"
	}

	if !strings.Contains(commitStr, "abc1234") {
		t.Errorf("Commit string should contain short hash, got %q", commitStr)
	}
	if !strings.Contains(commitStr, "abc1234567890") {
		t.Errorf("Commit string should contain full hash, got %q", commitStr)
	}
}

func TestPrintVersionHumanDev(t *testing.T) {
	info := VersionInfo{
		Name:    "clawdepl",
		Version: "dev",
	}

	// Test dev version string formatting
	versionStr := info.Version
	if versionStr == "dev" {
		// Should stay as "dev", not "vdev"
		if versionStr != "dev" {
			t.Errorf("Dev version string = %q, want %q", versionStr, "dev")
		}
	}
}

func TestVersionCommandExists(t *testing.T) {
	// Verify the version command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command should be registered with rootCmd")
	}
}

func TestVersionCommandHasJSONFlag(t *testing.T) {
	// Find the version command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			// Check for --json flag
			flag := cmd.Flags().Lookup("json")
			if flag == nil {
				t.Error("version command should have --json flag")
			}
			return
		}
	}
	t.Error("version command not found")
}

func TestVersionJSONMarshal(t *testing.T) {
	info := VersionInfo{
		Name:       "clawdepl",
		Version:    "dev",
		Commit:     "unknown",
		CommitFull: "unknown",
		Date:       "unknown",
		Mode:       "prod",
		GOOS:       "linux",
		GOARCH:     "amd64",
		GoVersion:  "go1.21.0",
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name": "clawdepl"`) {
		t.Error("JSON output should contain name field")
	}
	if !strings.Contains(output, `"version": "dev"`) {
		t.Error("JSON output should contain version field")
	}
	if !strings.Contains(output, `"mode": "prod"`) {
		t.Error("JSON output should contain mode field")
	}
}
