package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const anthropicVersion = "2023-06-01"

var anthropicModelsURL = "https://api.anthropic.com/v1/models"
var anthropicHTTPClient = &http.Client{Timeout: 20 * time.Second}

type anthropicErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ValidateAnthropicAPIKey validates an Anthropic API key before provisioning.
func ValidateAnthropicAPIKey(ctx context.Context, apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("Claude API token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, anthropicModelsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Anthropic validation request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := anthropicHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate Claude API token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	errMessage := parseAnthropicErrorMessage(body)

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		if errMessage != "" {
			return fmt.Errorf("invalid Claude API token: %s", errMessage)
		}
		return fmt.Errorf("invalid Claude API token")
	case http.StatusTooManyRequests:
		return fmt.Errorf("unable to validate Claude API token right now (rate limited by Anthropic)")
	default:
		if errMessage != "" {
			return fmt.Errorf("Claude API token validation failed (status %d): %s", resp.StatusCode, errMessage)
		}
		return fmt.Errorf("Claude API token validation failed (status %d)", resp.StatusCode)
	}
}

func parseAnthropicErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var parsed anthropicErrorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}

	return strings.TrimSpace(parsed.Error.Message)
}
