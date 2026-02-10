// Package api provides the client for interacting with the clawdepl hosted infrastructure API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
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
		APIURL: GetEffectiveAPIEndpoint(),
		// Prefer request-scoped context timeouts over a global http.Client.Timeout.
		Timeout: 0,
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
			Timeout:   cfg.Timeout,
			Transport: defaultHTTPTransport(),
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

func defaultHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
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

// CreateSandboxRequest represents a request to create a new sandbox.
//
// This matches the production create-sandbox schema consumed by the GUI.
type CreateSandboxRequest struct {
	MoltyName               string `json:"molty_name"`                // required
	AnthropicCredentialType string `json:"anthropic_credential_type"` // required: "api_key" or "token"
	AnthropicCredential     string `json:"anthropic_credential"`      // required
	MoltyPrompt             string `json:"molty_prompt"`              // required (IDENTITY.md content)
}

// CreateSandboxResponse represents the response from creating a sandbox.
//
// New schema returns {sandbox_id, gateway_auth_token}. We also keep legacy fields
// so older deployments can still be parsed without crashing the CLI.
type CreateSandboxResponse struct {
	SandboxID        string `json:"sandbox_id,omitempty"`
	GatewayAuthToken string `json:"gateway_auth_token,omitempty"`

	// Legacy fields (older response shape).
	Success bool   `json:"success,omitempty"`
	BotName string `json:"bot_name,omitempty"`
	Sandbox struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"sandbox,omitempty"`
	Error string `json:"error,omitempty"`
}

func (r *CreateSandboxResponse) EffectiveSandboxID() string {
	if r == nil {
		return ""
	}
	if strings.TrimSpace(r.SandboxID) != "" {
		return strings.TrimSpace(r.SandboxID)
	}
	return strings.TrimSpace(r.Sandbox.ID)
}

// SandboxExecRequest represents a request to the sandbox-exec endpoint
type SandboxExecRequest struct {
	SandboxID string `json:"sandbox_id"`
	Action    string `json:"action"`
	SessionID string `json:"session_id,omitempty"`
	Command   string `json:"command,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

// CreateSessionResponse represents the response from creating a sandbox terminal session.
type CreateSessionResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ExecResponse represents synchronous command execution output from sandbox-exec.
type ExecResponse struct {
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

// DeleteSessionResponse represents the response from deleting a sandbox terminal session.
type DeleteSessionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
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

// SSHTerminalRequest represents a request to the sandbox-terminal-url endpoint
type SSHTerminalRequest struct {
	SandboxID        string `json:"sandbox_id"`
	Action           string `json:"action"` // "create-ssh" or "revoke-ssh"
	ExpiresInMinutes int    `json:"expires_in_minutes,omitempty"`
	SSHToken         string `json:"ssh_token,omitempty"`
}

// CreateSSHResponse represents the response from creating SSH access
type CreateSSHResponse struct {
	SSHToken         string `json:"ssh_token"`
	SSHCommand       string `json:"ssh_command"` // "ssh token@ssh.app.daytona.io"
	ExpiresInMinutes int    `json:"expires_in_minutes"`
	SandboxID        string `json:"sandbox_id"`
}

// RevokeSSHResponse represents the response from revoking SSH access
type RevokeSSHResponse struct {
	Success   bool   `json:"success"`
	SandboxID string `json:"sandbox_id"`
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
func (c *Client) CreateSandbox(ctx context.Context, req *CreateSandboxRequest) (*CreateSandboxResponse, error) {
	url := fmt.Sprintf("%s/create-sandbox", c.apiURL)

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

// CreateSession creates a terminal session for exec commands via POST /sandbox-exec.
func (c *Client) CreateSession(ctx context.Context, sandboxID, sessionID string) (*CreateSessionResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "create-session",
		SessionID: sessionID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[CreateSessionResponse](resp)
}

// ExecCommand executes a command synchronously in a session via POST /sandbox-exec.
func (c *Client) ExecCommand(ctx context.Context, sandboxID, sessionID, command string, timeoutSeconds int) (*ExecResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "exec",
		SessionID: sessionID,
		Command:   command,
		Timeout:   timeoutSeconds,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[ExecResponse](resp)
}

// DeleteSession deletes a previously created terminal session via POST /sandbox-exec.
func (c *Client) DeleteSession(ctx context.Context, sandboxID, sessionID string) (*DeleteSessionResponse, error) {
	url := fmt.Sprintf("%s/sandbox-exec", c.apiURL)

	req := &SandboxExecRequest{
		SandboxID: sandboxID,
		Action:    "delete-session",
		SessionID: sessionID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[DeleteSessionResponse](resp)
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

// CreateSSH creates SSH access to a sandbox via POST /sandbox-terminal-url with action "create-ssh"
func (c *Client) CreateSSH(ctx context.Context, sandboxID string, expiresInMinutes int) (*CreateSSHResponse, error) {
	url := fmt.Sprintf("%s/sandbox-terminal-url", c.apiURL)

	req := &SSHTerminalRequest{
		SandboxID:        sandboxID,
		Action:           "create-ssh",
		ExpiresInMinutes: expiresInMinutes,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[CreateSSHResponse](resp)
}

// RevokeSSH revokes SSH access to a sandbox via POST /sandbox-terminal-url with action "revoke-ssh"
func (c *Client) RevokeSSH(ctx context.Context, sandboxID, sshToken string) (*RevokeSSHResponse, error) {
	url := fmt.Sprintf("%s/sandbox-terminal-url", c.apiURL)

	req := &SSHTerminalRequest{
		SandboxID: sandboxID,
		Action:    "revoke-ssh",
		SSHToken:  sshToken,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[RevokeSSHResponse](resp)
}
