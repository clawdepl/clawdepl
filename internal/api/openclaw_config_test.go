package api

import (
	"encoding/json"
	"testing"
)

func TestCanonicalAnthropicAuthChoice(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "", want: "anthropic-api-key"},
		{in: "api-key", want: "anthropic-api-key"},
		{in: "anthropic-api-key", want: "anthropic-api-key"},
		{in: "setup-token", want: "anthropic-setup-token"},
		{in: "token", want: "anthropic-setup-token"},
		{in: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		got, err := CanonicalAnthropicAuthChoice(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("expected error for %q", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("CanonicalAnthropicAuthChoice(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidateOpenClawModel(t *testing.T) {
	if got, err := ValidateOpenClawModel(""); err != nil || got != DefaultOpenClawModel {
		t.Fatalf("default model mismatch: got=%q err=%v", got, err)
	}

	if got, err := ValidateOpenClawModel("anthropic/claude-sonnet-4-5"); err != nil || got != "anthropic/claude-sonnet-4-5" {
		t.Fatalf("expected valid anthropic model, got=%q err=%v", got, err)
	}

	if got, err := ValidateOpenClawModel("z.ai/glm-4.5"); err != nil || got != "zai/glm-4.5" {
		t.Fatalf("expected z.ai alias normalization, got=%q err=%v", got, err)
	}

	if _, err := ValidateOpenClawModel("bad-format"); err == nil {
		t.Fatalf("expected error for bad format")
	}

	if _, err := ValidateOpenClawModel("unknown-provider/model"); err == nil {
		t.Fatalf("expected error for unknown provider")
	}
}

func TestBuildOpenClawConfig(t *testing.T) {
	raw, err := BuildOpenClawConfig("sk-ant-test", "setup-token", "anthropic/claude-sonnet-4-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	env := cfg["env"].(map[string]any)
	if env["ANTHROPIC_API_KEY"] != "sk-ant-test" {
		t.Fatalf("missing ANTHROPIC_API_KEY")
	}
	if env["ANTHROPIC_SETUP_TOKEN"] != "sk-ant-test" {
		t.Fatalf("missing ANTHROPIC_SETUP_TOKEN")
	}

	gateway := cfg["gateway"].(map[string]any)
	auth := gateway["auth"].(map[string]any)
	token := auth["token"].(string)
	if token == "" {
		t.Fatalf("missing gateway.auth.token")
	}
}

func TestExtractAnthropicCredentialSetupTokenWrapped(t *testing.T) {
	blob := "Your OAuth token:\n\nsk-ant-oat01-AAAA\nBBBB-CCCCQAA\n\nStore this token securely."
	got, ok := ExtractAnthropicCredential(blob, "anthropic-setup-token")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got != "sk-ant-oat01-AAAABBBB-CCCCQAA" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestExtractAnthropicCredentialSetupTokenWrappedWithSpacesOnBlankLine(t *testing.T) {
	blob := "Your OAuth token:\n\nsk-ant-oat01-AAAA\nBBBB-CCCCQAA\n  \nStore this token securely."
	got, ok := ExtractAnthropicCredential(blob, "anthropic-setup-token")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got != "sk-ant-oat01-AAAABBBB-CCCCQAA" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestExtractAnthropicCredentialSetupTokenNoSuffixAAA(t *testing.T) {
	blob := "sk-ant-oat01-fooBARbazQAA\n"
	got, ok := ExtractAnthropicCredential(blob, "anthropic-setup-token")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got != "sk-ant-oat01-fooBARbazQAA" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestExtractAnthropicCredentialSetupTokenWithTUIGlyphPrefixes(t *testing.T) {
	blob := "Enter your Claude credential:\n\n┃ sk-ant-oat01-AAAA\n┃ BBBB-CCCCQAA\n┃\n"
	got, ok := ExtractAnthropicCredential(blob, "anthropic-setup-token")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got != "sk-ant-oat01-AAAABBBB-CCCCQAA" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestExtractAnthropicCredentialSetupTokenSingleLineWithTrailingText(t *testing.T) {
	blob := "Your OAuth token: sk-ant-oat01-AAAABBBB-CCCCQAA Store this token securely."
	got, ok := ExtractAnthropicCredential(blob, "anthropic-setup-token")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got != "sk-ant-oat01-AAAABBBB-CCCCQAA" {
		t.Fatalf("unexpected token: %q", got)
	}
}
