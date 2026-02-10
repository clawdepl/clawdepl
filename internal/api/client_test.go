package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type clientRoundTripFunc func(*http.Request) (*http.Response, error)

func (f clientRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestCreateSandbox verifies the API client calls POST /create-sandbox correctly
func TestCreateSandbox(t *testing.T) {
	client := &Client{
		apiURL: "https://example.invalid",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: clientRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/create-sandbox" {
					t.Errorf("Expected path /create-sandbox, got %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST, got %s", r.Method)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header")
				}

				var body CreateSandboxRequest
				_ = json.NewDecoder(r.Body).Decode(&body)

				if body.MoltyName == "" {
					t.Errorf("Expected molty_name to be set")
				}
				if body.MoltyName != "test-agent" {
					t.Errorf("Expected molty_name 'test-agent', got %v", body.MoltyName)
				}
				if body.MoltyPrompt != "test-vibe" {
					t.Errorf("Expected molty_prompt 'test-vibe', got %v", body.MoltyPrompt)
				}
				if body.AnthropicCredentialType != "api_key" {
					t.Errorf("Expected anthropic_credential_type 'api_key', got %v", body.AnthropicCredentialType)
				}
				if body.AnthropicCredential != "sk-test" {
					t.Errorf("Expected anthropic_credential 'sk-test', got %v", body.AnthropicCredential)
				}

				response := CreateSandboxResponse{
					SandboxID:        "test-sandbox-123",
					GatewayAuthToken: "mv_deadbeef",
				}
				respBody, _ := json.Marshal(response)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Request:    r,
				}, nil
			}),
		},
		token:  "test-token",
		userID: "test-user",
	}

	resp, err := client.CreateSandbox(context.Background(), &CreateSandboxRequest{
		MoltyName:               "test-agent",
		MoltyPrompt:             "test-vibe",
		AnthropicCredentialType: "api_key",
		AnthropicCredential:     "sk-test",
	})
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}

	if resp.SandboxID != "test-sandbox-123" {
		t.Errorf("Expected sandbox ID 'test-sandbox-123', got %s", resp.SandboxID)
	}
}

// TestSandboxExecEndpoints verifies all sandbox-exec operations use POST /sandbox-exec
func TestSandboxExecEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		expectedAction string
		callFunc       func(*Client) error
	}{
		{
			name:           "CheckSandboxStatus",
			expectedAction: "check-sandbox-status",
			callFunc: func(c *Client) error {
				_, err := c.CheckSandboxStatus(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:           "StartSandbox",
			expectedAction: "start-sandbox",
			callFunc: func(c *Client) error {
				_, err := c.StartSandbox(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:           "StopSandbox",
			expectedAction: "stop-sandbox",
			callFunc: func(c *Client) error {
				_, err := c.StopSandbox(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:           "DeleteSandbox",
			expectedAction: "delete-sandbox",
			callFunc: func(c *Client) error {
				_, err := c.DeleteSandbox(context.Background(), "sandbox-123")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				apiURL: "https://example.invalid",
				httpClient: &http.Client{
					Timeout: 30 * time.Second,
					Transport: clientRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						if r.URL.Path != "/sandbox-exec" {
							t.Errorf("Expected path /sandbox-exec, got %s", r.URL.Path)
						}

						if r.Method != http.MethodPost {
							t.Errorf("Expected method POST, got %s", r.Method)
						}

						var body SandboxExecRequest
						_ = json.NewDecoder(r.Body).Decode(&body)

						if body.SandboxID != "sandbox-123" {
							t.Errorf("Expected sandbox_id 'sandbox-123', got %s", body.SandboxID)
						}

						if body.Action != tt.expectedAction {
							t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, body.Action)
						}

						respBody, _ := json.Marshal(map[string]any{
							"success": true,
							"state":   "running",
							"ready":   true,
						})

						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body:       io.NopCloser(strings.NewReader(string(respBody))),
							Request:    r,
						}, nil
					}),
				},
				token: "test-token",
			}

			if err := tt.callFunc(client); err != nil {
				t.Errorf("Call failed: %v", err)
			}
		})
	}
}
