package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	DefaultOpenClawModel = "anthropic/claude-opus-4-6"
)

var (
	modelRefPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*/[A-Za-z0-9._:/-]+$`)

	anthropicAPIKeyPattern = regexp.MustCompile(`sk-ant-api[0-9A-Za-z_-]+`)

	nativeOpenClawProviders = map[string]struct{}{
		"anthropic":          {},
		"openai":             {},
		"openai-codex":       {},
		"opencode":           {},
		"google":             {},
		"google-vertex":      {},
		"google-antigravity": {},
		"google-gemini-cli":  {},
		"zai":                {},
		"vercel-ai-gateway":  {},
		"openrouter":         {},
		"amazon-bedrock":     {},
		"xai":                {},
		"groq":               {},
		"cerebras":           {},
		"mistral":            {},
		"github-copilot":     {},
		"moonshot":           {},
		"kimi-coding":        {},
		"qwen-portal":        {},
		"qianfan":            {},
		"synthetic":          {},
		"minimax":            {},
		"ollama":             {},
	}
)

func CanonicalAnthropicAuthChoice(input string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", "api-key", "anthropic-api-key":
		return "anthropic-api-key", nil
	case "token", "setup-token", "anthropic-setup-token":
		return "anthropic-setup-token", nil
	default:
		return "", fmt.Errorf("invalid auth choice %q (use one of: api-key, setup-token)", input)
	}
}

// ValidateASCIICredential rejects credentials containing non-ASCII characters.
// This prevents hard-to-debug failures when users paste terminal output that includes
// unicode spinners/checkmarks alongside tokens.
func ValidateASCIICredential(credential string) error {
	s := strings.TrimSpace(credential)
	if s == "" {
		return fmt.Errorf("credential is required")
	}
	for i, r := range s {
		if r > 127 {
			return fmt.Errorf("credential contains a non-ASCII character at index %d (U+%04X). Paste only the raw token/key (no terminal output symbols)", i, r)
		}
	}
	return nil
}

func extractWrappedTokenByPrefix(raw string, prefix string) (string, bool) {
	// Tokens are ASCII and may be wrapped with newlines/spaces in CLI output.
	if strings.TrimSpace(raw) == "" {
		return "", false
	}

	start := strings.Index(raw, prefix)
	if start < 0 {
		return "", false
	}

	rest := raw[start:]
	// Prefer the first blank line as an end-of-token-block delimiter. This avoids accidentally
	// consuming explanatory text ("Store this token securely..."), which contains only ASCII
	// letters/digits that would otherwise look like token characters.
	cut := rest
	if idx := indexOfBlankLine(cut); idx >= 0 {
		cut = cut[:idx]
	}

	// The token itself can be wrapped across lines. Some UIs (including our TUI) may add
	// non-ASCII border glyphs (e.g. "┃") at the start of each wrapped line; treat those
	// as decoration and ignore them.
	//
	// Parse line-by-line and stop at the first line that isn't purely token characters
	// after stripping decoration. This works even when no blank line exists.
	isTokenChar := func(b byte) bool {
		return (b >= 'a' && b <= 'z') ||
			(b >= 'A' && b <= 'Z') ||
			(b >= '0' && b <= '9') ||
			b == '_' || b == '-'
	}
	stripLineDecoration := func(line string) string {
		s := strings.TrimLeft(line, " \t")
		// Common prefixes from terminal copy/paste and TUIs.
		for {
			trimmed := strings.TrimLeft(s, " \t")
			if trimmed == "" {
				return ""
			}
			// ASCII borders/prefixes.
			if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") {
				s = strings.TrimLeft(trimmed[1:], " \t")
				continue
			}
			// Unicode box drawing (most common).
			if strings.HasPrefix(trimmed, "┃") || strings.HasPrefix(trimmed, "│") {
				s = strings.TrimLeft(trimmed[len("┃"):], " \t")
				continue
			}
			return trimmed
		}
	}

	var out []byte
	lines := strings.Split(strings.ReplaceAll(cut, "\r\n", "\n"), "\n")
	for _, line := range lines {
		s := stripLineDecoration(line)
		if s == "" {
			// Empty line inside the cut: treat as end of token block.
			break
		}

		// Take the maximal prefix of token chars.
		i := 0
		for i < len(s) && isTokenChar(s[i]) {
			i++
		}
		if i == 0 {
			break
		}
		out = append(out, s[:i]...)
		// If the line contains anything other than token chars, stop after this segment.
		// This lets us handle single-line pastes like "sk-ant-oat... Store this token securely."
		if strings.TrimSpace(s[i:]) != "" {
			break
		}
	}

	tok := strings.TrimSpace(string(out))
	if !strings.HasPrefix(tok, prefix) {
		return "", false
	}
	// Sanity: avoid returning just the prefix.
	if len(tok) < len(prefix)+10 {
		return "", false
	}
	return tok, true
}

func indexOfBlankLine(s string) int {
	// Returns the index of the newline that begins a blank line (i.e. end of the token block),
	// or -1 if no blank line is found.
	//
	// Matches:
	//   "\n\n"
	//   "\n  \n"
	//   "\r\n\r\n"
	//   "\r\n \r\n"
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		// Find end-of-line
		if b[i] != '\n' && b[i] != '\r' {
			continue
		}
		j := i
		// Consume first line ending (\n or \r\n)
		if b[j] == '\r' {
			j++
			if j < len(b) && b[j] == '\n' {
				j++
			}
		} else {
			j++
		}
		// Skip spaces/tabs on the "blank" line
		for j < len(b) && (b[j] == ' ' || b[j] == '\t') {
			j++
		}
		// Must be another line ending immediately after optional whitespace
		if j >= len(b) {
			return -1
		}
		if b[j] != '\n' && b[j] != '\r' {
			continue
		}
		// Blank line found; return index of the first line ending (start of the blank separator)
		return i
	}
	return -1
}

// ExtractAnthropicCredential tries to extract the raw credential from a larger pasted blob
// (e.g. the full output of `claude setup-token` that includes unicode glyphs).
//
// This never returns non-ASCII characters. If no credential is found, it returns ("", false).
func ExtractAnthropicCredential(raw string, authMethod string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false
	}

	var m string
	switch authMethod {
	case "anthropic-setup-token":
		if tok, ok := extractWrappedTokenByPrefix(s, "sk-ant-oat"); ok {
			m = tok
		}
	case "anthropic-api-key":
		m = anthropicAPIKeyPattern.FindString(s)
		if m == "" {
			if tok, ok := extractWrappedTokenByPrefix(s, "sk-ant-api"); ok {
				m = tok
			}
		}
	default:
		// Unknown: try setup-token first, then api-key.
		if tok, ok := extractWrappedTokenByPrefix(s, "sk-ant-oat"); ok {
			m = tok
		}
		if m == "" {
			m = anthropicAPIKeyPattern.FindString(s)
			if m == "" {
				if tok, ok := extractWrappedTokenByPrefix(s, "sk-ant-api"); ok {
					m = tok
				}
			}
		}
	}

	if m == "" {
		return "", false
	}

	// Defensive: ensure ASCII-only match (regex already restricts, but keep invariant explicit).
	if err := ValidateASCIICredential(m); err != nil {
		return "", false
	}
	return m, true
}

func ValidateOpenClawModel(model string) (string, error) {
	m := strings.TrimSpace(model)
	if m == "" {
		m = DefaultOpenClawModel
	}

	if !modelRefPattern.MatchString(m) {
		return "", fmt.Errorf("invalid model %q (expected format: provider/model)", m)
	}

	parts := strings.SplitN(m, "/", 2)
	provider := normalizeProviderAlias(parts[0])
	if _, ok := nativeOpenClawProviders[provider]; !ok {
		return "", fmt.Errorf("provider %q is not in OpenClaw native providers", provider)
	}

	return provider + "/" + parts[1], nil
}

func BuildOpenClawConfig(anthropicCredential, authChoice, model string) (string, error) {
	credential := strings.TrimSpace(anthropicCredential)
	if credential == "" {
		return "", fmt.Errorf("Claude credential is required")
	}

	choice, err := CanonicalAnthropicAuthChoice(authChoice)
	if err != nil {
		return "", err
	}

	modelRef, err := ValidateOpenClawModel(model)
	if err != nil {
		return "", err
	}

	gatewayToken, err := generateGatewayToken()
	if err != nil {
		return "", err
	}

	env := map[string]string{
		"ANTHROPIC_API_KEY": credential,
	}
	if choice == "anthropic-setup-token" {
		// Setup-token is a supported Anthropic auth mode in OpenClaw.
		// Keep API-key field for compatibility and provide explicit token field.
		env["ANTHROPIC_SETUP_TOKEN"] = credential
	}

	models := map[string]any{
		"providers": map[string]any{
			"anthropic": map[string]any{
				// Keep schema compatible with the snapshot image in openclaw-docker.
				// This is redundant with env, but harmless and makes config more portable.
				"apiKey": credential,
			},
		},
	}
	if choice == "anthropic-setup-token" {
		models["providers"].(map[string]any)["anthropic"].(map[string]any)["setupToken"] = credential
	}

	config := map[string]any{
		"env": env,
		"agents": map[string]any{
			"defaults": map[string]any{
				"model": map[string]string{
					"primary": modelRef,
				},
			},
		},
		"gateway": map[string]any{
			"mode": "local",
			"auth": map[string]any{
				"token": gatewayToken,
			},
		},
		"models": models,
		"clawdepl": map[string]any{
			"auth": map[string]string{
				"provider": "anthropic",
				"method":   choice,
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to build OpenClaw config: %w", err)
	}

	return string(data), nil
}

func generateGatewayToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate gateway token: %w", err)
	}
	// URL-safe string; no padding.
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func normalizeProviderAlias(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "z.ai", "z-ai":
		return "zai"
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}
