package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestProvisionAsync verifies the API client calls Convex HTTP endpoint correctly
func TestProvisionAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's calling the HTTP endpoint, not action API
		if r.URL.Path != "/api/provision" {
			t.Errorf("Expected path /api/provision, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header")
		}

		var body ProvisionRequest
		json.NewDecoder(r.Body).Decode(&body)

		if body.MoltyName != "test-agent" {
			t.Errorf("Expected moltyName 'test-agent', got %v", body.MoltyName)
		}

		// Send standard HTTP response
		response := ProvisionResponse{
			Success:   true,
			SandboxID: "test-sandbox-123",
			Status:    "provisioning",
			Message:   "Sandbox created",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		convexURL:  server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      "test-token",
		userID:     "test-user",
	}

	resp, err := client.ProvisionAsync(context.Background(), "test-agent", "sk-test", "test-vibe")
	if err != nil {
		t.Fatalf("ProvisionAsync failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false")
	}

	if resp.SandboxID != "test-sandbox-123" {
		t.Errorf("Expected sandboxId 'test-sandbox-123', got %s", resp.SandboxID)
	}
}

// TestConvexHTTPEndpoints verifies all operations call HTTP endpoints correctly
func TestConvexHTTPEndpoints(t *testing.T) {
	tests := []struct {
		name         string
		expectedPath string
		expectedMethod string
		callFunc     func(*Client) error
	}{
		{
			name:         "GetStatus",
			expectedPath: "/api/provision/status/sandbox-123",
			expectedMethod: "GET",
			callFunc: func(c *Client) error {
				_, err := c.GetStatus(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:         "StartSandbox",
			expectedPath: "/api/provision/start/sandbox-123",
			expectedMethod: "POST",
			callFunc: func(c *Client) error {
				_, err := c.StartSandbox(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:         "StopSandbox",
			expectedPath: "/api/provision/stop/sandbox-123",
			expectedMethod: "POST",
			callFunc: func(c *Client) error {
				_, err := c.StopSandbox(context.Background(), "sandbox-123")
				return err
			},
		},
		{
			name:         "Deprovision",
			expectedPath: "/api/provision/deprovision",
			expectedMethod: "POST",
			callFunc: func(c *Client) error {
				_, err := c.Deprovision(context.Background(), "sandbox-123")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}

				if r.Method != tt.expectedMethod {
					t.Errorf("Expected method %s, got %s", tt.expectedMethod, r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client := &Client{
				convexURL:  server.URL,
				httpClient: &http.Client{Timeout: 30 * time.Second},
				token:      "test-token",
			}

			if err := tt.callFunc(client); err != nil {
				t.Errorf("Call failed: %v", err)
			}
		})
	}
}
