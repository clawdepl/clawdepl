package cmd

import (
	"strings"
	"testing"
)

func TestIsNonInteractiveNewMode(t *testing.T) {
	origToken := newClaudeToken
	origPurpose := newPurpose
	t.Cleanup(func() {
		newClaudeToken = origToken
		newPurpose = origPurpose
	})

	newClaudeToken = ""
	newPurpose = ""
	if isNonInteractiveNewMode() {
		t.Fatal("expected interactive mode when both flags are empty")
	}

	newClaudeToken = "sk-ant-api03-test"
	newPurpose = ""
	if !isNonInteractiveNewMode() {
		t.Fatal("expected non-interactive mode when --claude-token is set")
	}

	newClaudeToken = ""
	newPurpose = "test purpose"
	if !isNonInteractiveNewMode() {
		t.Fatal("expected non-interactive mode when --purpose is set")
	}
}

func TestRunNewNonInteractiveValidation(t *testing.T) {
	origToken := newClaudeToken
	origPurpose := newPurpose
	origAuthChoice := newAuthChoice
	origModel := newModel
	t.Cleanup(func() {
		newClaudeToken = origToken
		newPurpose = origPurpose
		newAuthChoice = origAuthChoice
		newModel = origModel
	})

	newClaudeToken = "sk-ant-api03-test"
	newPurpose = "test purpose"
	newAuthChoice = "api-key"
	newModel = "anthropic/claude-opus-4-6"
	err := runNewNonInteractive(nil, "")
	if err == nil || !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected name validation error, got: %v", err)
	}

	newClaudeToken = ""
	newPurpose = "test purpose"
	err = runNewNonInteractive(nil, "test-name")
	if err == nil || !strings.Contains(err.Error(), "--claude-token is required") {
		t.Fatalf("expected token validation error, got: %v", err)
	}

	newClaudeToken = "sk-ant-api03-test"
	newPurpose = ""
	err = runNewNonInteractive(nil, "test-name")
	if err == nil || !strings.Contains(err.Error(), "--purpose is required") {
		t.Fatalf("expected purpose validation error, got: %v", err)
	}
}
