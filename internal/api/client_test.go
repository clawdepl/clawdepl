package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCreateSandbox verifies the API client calls POST /create-sandbox correctly
func TestCreateSandbox(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		json.NewDecoder(r.Body).Decode(&body)

		if body.Name != "test-agent" {
			t.Errorf("Expected name 'test-agent', got %v", body.Name)
		}
		if body.OpenclawConfig != "sk-test" {
			t.Errorf("Expected openclaw_config 'sk-test', got %v", body.OpenclawConfig)
		}
		if body.MoltyPrompt != "test-vibe" {
			t.Errorf("Expected molty_prompt 'test-vibe', got %v", body.MoltyPrompt)
		}

		response := CreateSandboxResponse{
			Success: true,
			BotName: "test-agent",
		}
		response.Sandbox.ID = "test-sandbox-123"
		response.Sandbox.State = "provisioning"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiURL:     server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      "test-token",
		userID:     "test-user",
	}

	resp, err := client.CreateSandbox(context.Background(), "test-agent", "sk-test", "test-vibe")
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	if resp.Sandbox.ID != "test-sandbox-123" {
		t.Errorf("Expected sandbox ID 'test-sandbox-123', got %s", resp.Sandbox.ID)
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/sandbox-exec" {
					t.Errorf("Expected path /sandbox-exec, got %s", r.URL.Path)
				}

				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				var body SandboxExecRequest
				json.NewDecoder(r.Body).Decode(&body)

				if body.SandboxID != "sandbox-123" {
					t.Errorf("Expected sandbox_id 'sandbox-123', got %s", body.SandboxID)
				}

				if body.Action != tt.expectedAction {
					t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, body.Action)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"state":   "running",
					"ready":   true,
				})
			}))
			defer server.Close()

			client := &Client{
				apiURL:     server.URL,
				httpClient: &http.Client{Timeout: 30 * time.Second},
				token:      "test-token",
			}

			if err := tt.callFunc(client); err != nil {
				t.Errorf("Call failed: %v", err)
			}
		})
	}
}
