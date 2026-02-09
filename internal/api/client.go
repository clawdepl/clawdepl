// Package api provides the client for interacting with the clawdepl hosted infrastructure API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/clawdepl/clawdepl/internal/config"
)

// Client represents the clawdepl API client
type Client struct {
	convexURL       string
	provisionerURL  string
	httpClient      *http.Client
	token           string
	userID          string
}

// ClientConfig holds the configuration for the API client
type ClientConfig struct {
	ConvexURL      string
	ProvisionerURL string
	Token          string
	UserID         string
	Timeout        time.Duration
}

// Endpoint overrides for debug builds
var (
	provisionerEndpointOverride string
	convexEndpointOverride      string
	authEndpointOverride        string
	tokenOverride               string
	userIDOverride              string
)

// SetProvisionerEndpointOverride sets the provisioner endpoint override (called from cmd package in debug builds)
func SetProvisionerEndpointOverride(endpoint string) {
	provisionerEndpointOverride = endpoint
}

// SetConvexEndpointOverride sets the Convex endpoint override (called from cmd package in debug builds)
func SetConvexEndpointOverride(endpoint string) {
	convexEndpointOverride = endpoint
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

// GetEffectiveProvisionerEndpoint returns the provisioner endpoint to use
func GetEffectiveProvisionerEndpoint() string {
	if provisionerEndpointOverride != "" {
		return provisionerEndpointOverride
	}
	return buildinfo.ProvisionerEndpoint
}

// GetEffectiveConvexEndpoint returns the Convex endpoint to use
func GetEffectiveConvexEndpoint() string {
	if convexEndpointOverride != "" {
		return convexEndpointOverride
	}
	return buildinfo.ConvexEndpoint
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
		ConvexURL:      GetEffectiveConvexEndpoint(),
		ProvisionerURL: GetEffectiveProvisionerEndpoint(),
		Timeout:        30 * time.Second,
	}
}

// NewClient creates a new clawdepl API client
func NewClient(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	client := &Client{
		convexURL:      cfg.ConvexURL,
		provisionerURL: cfg.ProvisionerURL,
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

// Molty represents a Molty (AI agent) from the Convex backend
type Molty struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	SandboxID  string  `json:"sandboxId,omitempty"`
	GatewayURL string  `json:"gatewayUrl,omitempty"`
	CreatedAt  float64 `json:"createdAt"`
	OwnerID    string  `json:"ownerId,omitempty"`
}

// ProvisionRequest represents a request to provision a new sandbox
type ProvisionRequest struct {
	UserID      string      `json:"userId"`
	MoltyName   string      `json:"moltyName"`
	APIKey      string      `json:"apiKey"`
	Personality Personality `json:"personality,omitempty"`
}

// Personality represents the personality configuration for a Molty
type Personality struct {
	Name string `json:"name,omitempty"`
	Vibe string `json:"vibe,omitempty"`
}

// ProvisionResponse represents the response from provisioning
type ProvisionResponse struct {
	Success   bool   `json:"success"`
	SandboxID string `json:"sandboxId"`
	AuthToken string `json:"authToken,omitempty"`
	Status    string `json:"status"`
	Stage     string `json:"stage,omitempty"`
	Message   string `json:"message,omitempty"`
	Progress  int    `json:"progress,omitempty"`
}

// StatusResponse represents the status of a sandbox
type StatusResponse struct {
	SandboxID  string `json:"sandboxId"`
	MoltyName  string `json:"moltyName,omitempty"`
	Status     string `json:"status"`
	Stage      string `json:"stage,omitempty"`
	Message    string `json:"message,omitempty"`
	Progress   int    `json:"progress,omitempty"`
	GatewayURL string `json:"gatewayUrl,omitempty"`
	AuthToken  string `json:"authToken,omitempty"`
}

// DeprovisionRequest represents a request to deprovision a sandbox
type DeprovisionRequest struct {
	SandboxID string `json:"sandboxId"`
	UserID    string `json:"userId"`
}

// DeprovisionResponse represents the response from deprovisioning
type DeprovisionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ActionResponse represents a generic action response (start/stop/restart)
type ActionResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// MoltysResponse represents the response from listing moltys
type MoltysResponse struct {
	Moltys []Molty `json:"moltys"`
}

// ========== Convex API Methods ==========

// ListMoltys lists all moltys for the authenticated user (via Convex query API)
func (c *Client) ListMoltys(ctx context.Context) ([]Molty, error) {
	// Use Convex's public query API endpoint (different from HTTP routes)
	// The .site URL is for HTTP routes, .cloud URL is for direct query calls
	convexCloudURL := strings.Replace(c.convexURL, ".convex.site", ".convex.cloud", 1)
	url := fmt.Sprintf("%s/api/query", convexCloudURL)

	// Convex query API format: POST with {"path":"module:function","args":{}}
	queryBody := map[string]interface{}{
		"path": "moltys:list",
		"args": map[string]interface{}{},
	}
	jsonBody, err := json.Marshal(queryBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Convex query API returns {"status":"success","value":[...]}
	var queryResponse struct {
		Status string  `json:"status"`
		Value  []Molty `json:"value"`
	}
	if err := json.Unmarshal(body, &queryResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if queryResponse.Status != "success" {
		return nil, fmt.Errorf("query failed with status: %s", queryResponse.Status)
	}

	// Filter by ownerId client-side
	var userMoltys []Molty
	for _, m := range queryResponse.Value {
		if m.OwnerID == c.userID {
			userMoltys = append(userMoltys, m)
		}
	}

	return userMoltys, nil
}

// ========== Provisioner API Methods ==========

// ProvisionAsync starts async provisioning of a new sandbox
func (c *Client) ProvisionAsync(ctx context.Context, name, apiKey, vibe string) (*ProvisionResponse, error) {
	url := fmt.Sprintf("%s/api/provision/async", c.provisionerURL)

	req := &ProvisionRequest{
		UserID:    c.userID,
		MoltyName: name,
		APIKey:    apiKey,
		Personality: Personality{
			Name: name,
			Vibe: vibe,
		},
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[ProvisionResponse](resp)
}

// GetStatus gets the status of a sandbox
func (c *Client) GetStatus(ctx context.Context, sandboxID string) (*StatusResponse, error) {
	url := fmt.Sprintf("%s/api/status/%s", c.provisionerURL, sandboxID)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[StatusResponse](resp)
}

// StartSandbox starts a stopped sandbox
func (c *Client) StartSandbox(ctx context.Context, sandboxID string) (*ActionResponse, error) {
	url := fmt.Sprintf("%s/api/start/%s", c.provisionerURL, sandboxID)

	resp, err := c.doRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[ActionResponse](resp)
}

// StopSandbox stops a running sandbox
func (c *Client) StopSandbox(ctx context.Context, sandboxID string) (*ActionResponse, error) {
	url := fmt.Sprintf("%s/api/stop/%s", c.provisionerURL, sandboxID)

	resp, err := c.doRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[ActionResponse](resp)
}

// Deprovision deletes a sandbox
func (c *Client) Deprovision(ctx context.Context, sandboxID string) (*DeprovisionResponse, error) {
	url := fmt.Sprintf("%s/api/deprovision", c.provisionerURL)

	req := &DeprovisionRequest{
		SandboxID: sandboxID,
		UserID:    c.userID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	return parseResponse[DeprovisionResponse](resp)
}

// HealthCheck verifies connectivity to the provisioner API
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.provisionerURL)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// WaitForProvisioning polls the status endpoint until provisioning is complete
func (c *Client) WaitForProvisioning(ctx context.Context, sandboxID string, onProgress func(stage string, progress int, message string)) (*StatusResponse, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := c.GetStatus(ctx, sandboxID)
			if err != nil {
				return nil, err
			}

			if onProgress != nil {
				onProgress(status.Stage, status.Progress, status.Message)
			}

			// Check if provisioning is complete
			switch status.Status {
			case "running", "ready":
				return status, nil
			case "failed", "error":
				return nil, fmt.Errorf("provisioning failed: %s", status.Message)
			}
			// Continue polling for "provisioning" status
		}
	}
}
