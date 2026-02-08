// Package api provides the client for interacting with the OpenClaw hosted infrastructure API.
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client represents the OpenClaw API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Config holds the configuration for the API client
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// DefaultConfig returns the default API client configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.openclaw.io",
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new OpenClaw API client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Instance represents an OpenClaw instance
type Instance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateInstanceRequest represents a request to create a new instance
type CreateInstanceRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

// CreateInstance creates a new OpenClaw instance
// TODO: Implement when API spec is available
func (c *Client) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error) {
	return nil, fmt.Errorf("not implemented: API endpoint not yet available")
}

// GetInstance retrieves an instance by ID
// TODO: Implement when API spec is available
func (c *Client) GetInstance(ctx context.Context, id string) (*Instance, error) {
	return nil, fmt.Errorf("not implemented: API endpoint not yet available")
}

// ListInstances lists all instances for the authenticated user
// TODO: Implement when API spec is available
func (c *Client) ListInstances(ctx context.Context) ([]*Instance, error) {
	return nil, fmt.Errorf("not implemented: API endpoint not yet available")
}

// DeleteInstance deletes an instance by ID
// TODO: Implement when API spec is available
func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented: API endpoint not yet available")
}

// HealthCheck verifies connectivity to the API
// TODO: Implement when API spec is available
func (c *Client) HealthCheck(ctx context.Context) error {
	return fmt.Errorf("not implemented: API endpoint not yet available")
}
