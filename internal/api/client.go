// Package api provides the client for interacting with the clawdepl hosted infrastructure API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/clawdepl/clawdepl/internal/config"
)

// Client represents the clawdepl API client
type Client struct {
	apiURL     string
	httpClient *http.Client
	token      string
	userID     string
}

// ClientConfig holds the configuration for the API client
type ClientConfig struct {
	APIURL  string
	Token   string
	UserID  string
	Timeout time.Duration
}

// Endpoint overrides for debug builds
var (
	apiEndpointOverride  string
	authEndpointOverride string
	tokenOverride        string
	userIDOverride       string
)

// SetAPIEndpointOverride sets the API endpoint override (called from cmd package in debug builds)
func SetAPIEndpointOverride(endpoint string) {
	apiEndpointOverride = endpoint
}

// SetAuthEndpointOverride sets the auth endpoint override (called from cmd package in debug builds)
func SetAuthEndpointOverride(endpoint string) {
	authEndpointOverride = endpoint
}

// SetTokenOverride sets the auth token override (called from cmd package in debug builds)
func SetTokenOverride(token string) {
	tokenOverride = token
}

// SetUserIDOverride sets the user ID override (called from cmd package in debug builds)
func SetUserIDOverride(userID string) {
	userIDOverride = userID
}

// GetEffectiveAPIEndpoint returns the API endpoint to use
func GetEffectiveAPIEndpoint() string {
	if apiEndpointOverride != "" {
		return apiEndpointOverride
	}
	return buildinfo.APIEndpoint
}

// GetEffectiveAuthEndpoint returns the auth endpoint to use
func GetEffectiveAuthEndpoint() string {
	if authEndpointOverride != "" {
		return authEndpointOverride
	}
	return buildinfo.AuthEndpoint
}

// HasTokenOverride returns true if a token override is set
func HasTokenOverride() bool {
	return tokenOverride != ""
}

// GetTokenOverride returns the token override value
func GetTokenOverride() string {
	return tokenOverride
}

// GetUserIDOverride returns the user ID override value
func GetUserIDOverride() string {
	if userIDOverride != "" {
		return userIDOverride
	}
	return "debug-user"
}

// DefaultConfig returns the default API client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		APIURL:  GetEffectiveAPIEndpoint(),
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new clawdepl API client
func NewClient(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	client := &Client{
		apiURL: cfg.APIURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	// Check for token override first (debug builds)
	if HasTokenOverride() {
		client.token = GetTokenOverride()
		client.userID = GetUserIDOverride()
		return client, nil
	}

	// Otherwise, load from credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds != nil {
		client.token = creds.AccessToken
		if creds.User != nil {
			client.userID = creds.User.ID
		}
	}

	return client, nil
}

// doRequest performs an HTTP request with auth headers
func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// parseResponse parses the JSON response body
func parseResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ========== Data Types ==========

// CreateSandboxRequest represents a request to create a new sandbox
type CreateSandboxRequest struct {
	Name           string `json:"name"`
	OpenclawConfig string `json:"openclaw_config"`
	MoltyPrompt    string `json:"molty_prompt"`
}

// CreateSandboxResponse represents the response from creating a sandbox
type CreateSandboxResponse struct {
	Success bool   `json:"success"`
	BotName string `json:"bot_name,omitempty"`
	Sandbox struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"sandbox,omitempty"`
	Error string `json:"error,omitempty"`
}

// SandboxExecRequest represents a request to the sandbox-exec endpoint
type SandboxExecRequest struct {
	SandboxID string `json:"sandbox_id"`
	Action    string `json:"action"`
	SessionID string `json:"session_id,omitempty"`
	Command   string `json:"command,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

// SandboxStatusResponse represents the status check response
type SandboxStatusResponse struct {
	State string `json:"state"`
	Ready bool   `json:"ready"`
}

// ActionResponse represents a generic action response (start/stop/delete)
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Bot represents a bot returned by the list-bots endpoint
type Bot struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SandboxID string `json:"sandbox_id"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

// ListBotsResponse represents the response from listing bots
type ListBotsResponse struct {
	Bots []Bot `json:"bots"`
}

// ========== API Methods ==========

// ListBots returns all bots for the authenticated user via POST /list-bots
func (c *Client) ListBots(ctx context.Context) ([]Bot, error) {
	url := fmt.Sprintf("%s/list-bots", c.apiURL)

	resp, err := c.doRequest(ctx, http.MethodPost, url, map[string]string{})
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[ListBotsResponse](resp)
	if err != nil {
		return nil, err
	}

	return result.Bots, nil
}

// CreateSandbox creates a new sandbox via POST /create-sandbox
func (c *Client) CreateSandbox(ctx context.Context, name, openclawConfig, moltyPrompt string) (*CreateSandboxResponse, error) {
	url := fmt.Sprintf("%s/create-sandbox", c.apiURL)

	req := &CreateSandboxRequest{
		Name:           name,
		OpenclawConfig: openclawConfig,
		MoltyPrompt:    moltyPrompt,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[CreateSandboxResponse](resp)
}

// CheckSandboxStatus checks the status of a sandbox via POST /sandbox-exec with action "check-sandbox-status"
func (c *Client) CheckSandboxStatus(ctx context.Context, sandboxID string) (*SandboxStatusResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "check-sandbox-status",
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[SandboxStatusResponse](resp)
}

// StartSandbox starts a stopped sandbox via POST /sandbox-exec with action "start-sandbox"
func (c *Client) StartSandbox(ctx context.Context, sandboxID string) (*ActionResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "start-sandbox",
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[ActionResponse](resp)
}

// StopSandbox stops a running sandbox via POST /sandbox-exec with action "stop-sandbox"
func (c *Client) StopSandbox(ctx context.Context, sandboxID string) (*ActionResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "stop-sandbox",
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[ActionResponse](resp)
}

// DeleteSandbox deletes a sandbox via POST /sandbox-exec with action "delete-sandbox"
func (c *Client) DeleteSandbox(ctx context.Context, sandboxID string) (*ActionResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "delete-sandbox",
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[ActionResponse](resp)
}

// WaitForReady polls CheckSandboxStatus every 2s until the sandbox is ready
func (c *Client) WaitForReady(ctx context.Context, sandboxID string, onProgress func(state string, ready bool)) (*SandboxStatusResponse, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := c.CheckSandboxStatus(ctx, sandboxID)
			if err != nil {
				return nil, err
			}

			if onProgress != nil {
				onProgress(status.State, status.Ready)
			}

			if status.Ready {
				return status, nil
			}

			// Check for terminal failure states
			if status.State == "failed" || status.State == "error" {
				return nil, fmt.Errorf("sandbox entered %s state", status.State)
			}
		}
	}
}
