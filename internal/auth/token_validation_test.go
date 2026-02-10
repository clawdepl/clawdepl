package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/clawdepl/clawdepl/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLoginWithTokenRejectsInvalidToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	originalClient := validateHTTPClient
	t.Cleanup(func() { validateHTTPClient = originalClient })

	validateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/verify-token" {
				t.Fatalf("expected /verify-token, got %s", r.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"valid":false,"error":"Invalid or revoked token"}`)),
				Request:    r,
			}, nil
		}),
	}

	original := buildinfo.APIEndpoint
	buildinfo.APIEndpoint = "https://example.invalid"
	t.Cleanup(func() {
		buildinfo.APIEndpoint = original
	})

	_, err := LoginWithToken(context.Background(), "bad-token")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}

	creds, loadErr := config.LoadCredentials()
	if loadErr != nil {
		t.Fatalf("unexpected load error: %v", loadErr)
	}
	if creds != nil {
		t.Fatalf("credentials should not be saved on failed validation")
	}
}

func TestLoginWithAPIKeyValidatesAndStores(t *testing.T) {
	originalClient := validateHTTPClient
	t.Cleanup(func() { validateHTTPClient = originalClient })

	validateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/verify-token" {
				t.Fatalf("expected /verify-token, got %s", r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer api-key-123" {
				t.Fatalf("unexpected authorization header: %s", got)
			}
			body, _ := json.Marshal(map[string]any{
				"valid":  true,
				"userId": "u_123",
				"email":  "user@example.com",
				"name":   "User",
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(body))),
				Request:    r,
			}, nil
		}),
	}

	original := buildinfo.APIEndpoint
	buildinfo.APIEndpoint = "https://example.invalid"
	t.Cleanup(func() {
		buildinfo.APIEndpoint = original
	})

	creds, err := loginWithAPIKeyValue(context.Background(), "api-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.APIKey != "api-key-123" {
		t.Fatalf("expected API key to be saved")
	}
	if creds.User == nil || creds.User.ID != "u_123" {
		t.Fatalf("expected validated user details to be saved")
	}
}
