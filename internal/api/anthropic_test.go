package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestValidateAnthropicAPIKey(t *testing.T) {
	originalURL := anthropicModelsURL
	originalClient := anthropicHTTPClient
	t.Cleanup(func() {
		anthropicModelsURL = originalURL
		anthropicHTTPClient = originalClient
	})

	t.Run("valid key", func(t *testing.T) {
		anthropicModelsURL = "https://example.invalid/v1/models"
		anthropicHTTPClient = &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodGet {
					t.Fatalf("expected GET, got %s", r.Method)
				}
				if r.Header.Get("x-api-key") != "valid-key" {
					t.Fatalf("expected x-api-key header")
				}
				if r.Header.Get("anthropic-version") != anthropicVersion {
					t.Fatalf("expected anthropic-version header")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
					Request:    r,
				}, nil
			}),
		}
		if err := ValidateAnthropicAPIKey(context.Background(), "valid-key"); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		if err := ValidateAnthropicAPIKey(context.Background(), ""); err == nil {
			t.Fatalf("expected error for empty key")
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		anthropicModelsURL = "https://example.invalid/v1/models"
		anthropicHTTPClient = &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"error":{"type":"authentication_error","message":"invalid x-api-key"}}`)),
					Request:    r,
				}, nil
			}),
		}
		err := ValidateAnthropicAPIKey(context.Background(), "bad-key")
		if err == nil {
			t.Fatalf("expected error")
		}
		if got := err.Error(); got != "invalid Claude API token: invalid x-api-key" {
			t.Fatalf("unexpected error message: %s", got)
		}
	})
}
