// Package api provides the client for interacting with the OpenClaw hosted infrastructure API.
package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/moltyverse/clawdpl/internal/buildinfo"
	"github.com/moltyverse/clawdpl/internal/config"
)

// Client represents the OpenClaw API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	creds      *config.Credentials
}

// Config holds the configuration for the API client
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// endpointOverride is set by cmd package for debug builds
var endpointOverride string

// SetEndpointOverride sets the endpoint override (called from cmd package in debug builds)
func SetEndpointOverride(endpoint string) {
	endpointOverride = endpoint
}

// GetEffectiveEndpoint returns the API endpoint to use.
// Priority: endpointOverride (debug builds) > buildinfo.DefaultEndpoint
func GetEffectiveEndpoint() string {
	if endpointOverride != "" {
		return endpointOverride
	}
	return buildinfo.DefaultEndpoint
}

// DefaultConfig returns the default API client configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL: GetEffectiveEndpoint(),
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new OpenClaw API client
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return &Client{
		baseURL: cfg.BaseURL,
		creds:   creds,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// NewClientWithCredentials creates a new client with specific credentials
func NewClientWithCredentials(cfg *Config, creds *config.Credentials) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		baseURL: cfg.BaseURL,
		creds:   creds,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Instance represents an OpenClaw instance
type Instance struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Region      string    `json:"region"`
	Purpose     string    `json:"purpose,omitempty"`
	ClaudeToken string    `json:"claude_token,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
}

// CreateInstanceRequest represents a request to create a new instance
type CreateInstanceRequest struct {
	Name        string `json:"name"`
	ClaudeToken string `json:"claude_token"`
	Purpose     string `json:"purpose"`
	Region      string `json:"region,omitempty"`
}

// Mock data store for development
var (
	mockInstances = make(map[string]*Instance)
	mockMutex     sync.RWMutex
)

// initMockData initializes some mock instances for testing
func initMockData() {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	if len(mockInstances) == 0 {
		// Add some sample instances
		mockInstances["demo-bot"] = &Instance{
			ID:        "inst_abc123",
			Name:      "demo-bot",
			Status:    "running",
			Region:    "us-east-1",
			Purpose:   "Customer support automation",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
			StartedAt: time.Now().Add(-24 * time.Hour),
		}
		mockInstances["test-agent"] = &Instance{
			ID:        "inst_def456",
			Name:      "test-agent",
			Status:    "stopped",
			Region:    "eu-west-1",
			Purpose:   "Testing and development",
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		}
	}
}

// CreateInstance creates a new OpenClaw instance
func (c *Client) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error) {
	initMockData()
	mockMutex.Lock()
	defer mockMutex.Unlock()

	// Check if name already exists
	if _, exists := mockInstances[req.Name]; exists {
		return nil, fmt.Errorf("instance '%s' already exists", req.Name)
	}

	// Create mock instance
	instance := &Instance{
		ID:          fmt.Sprintf("inst_%d", time.Now().UnixNano()%1000000),
		Name:        req.Name,
		Status:      "provisioning",
		Region:      req.Region,
		Purpose:     req.Purpose,
		ClaudeToken: req.ClaudeToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if instance.Region == "" {
		instance.Region = "us-east-1"
	}

	mockInstances[req.Name] = instance
	return instance, nil
}

// GetInstance retrieves an instance by name
func (c *Client) GetInstance(ctx context.Context, name string) (*Instance, error) {
	initMockData()
	mockMutex.RLock()
	defer mockMutex.RUnlock()

	instance, exists := mockInstances[name]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", name)
	}

	return instance, nil
}

// ListInstances lists all instances for the authenticated user
func (c *Client) ListInstances(ctx context.Context) ([]*Instance, error) {
	initMockData()
	mockMutex.RLock()
	defer mockMutex.RUnlock()

	instances := make([]*Instance, 0, len(mockInstances))
	for _, inst := range mockInstances {
		instances = append(instances, inst)
	}

	return instances, nil
}

// DeleteInstance deletes an instance by name
func (c *Client) DeleteInstance(ctx context.Context, name string) error {
	initMockData()
	mockMutex.Lock()
	defer mockMutex.Unlock()

	if _, exists := mockInstances[name]; !exists {
		return fmt.Errorf("instance '%s' not found", name)
	}

	delete(mockInstances, name)
	return nil
}

// StartInstance starts an instance
func (c *Client) StartInstance(ctx context.Context, name string) (*Instance, error) {
	initMockData()
	mockMutex.Lock()
	defer mockMutex.Unlock()

	instance, exists := mockInstances[name]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", name)
	}

	if instance.Status == "running" {
		return nil, fmt.Errorf("instance '%s' is already running", name)
	}

	instance.Status = "running"
	instance.StartedAt = time.Now()
	instance.UpdatedAt = time.Now()

	return instance, nil
}

// StopInstance stops an instance
func (c *Client) StopInstance(ctx context.Context, name string) (*Instance, error) {
	initMockData()
	mockMutex.Lock()
	defer mockMutex.Unlock()

	instance, exists := mockInstances[name]
	if !exists {
		return nil, fmt.Errorf("instance '%s' not found", name)
	}

	if instance.Status == "stopped" {
		return nil, fmt.Errorf("instance '%s' is already stopped", name)
	}

	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()

	return instance, nil
}

// HealthCheck verifies connectivity to the API
func (c *Client) HealthCheck(ctx context.Context) error {
	// Mock: always healthy
	return nil
}

// WaitForProvisioning waits for an instance to be provisioned (mock: 3 seconds)
func (c *Client) WaitForProvisioning(ctx context.Context, name string) (*Instance, error) {
	// Simulate provisioning delay
	select {
	case <-time.After(3 * time.Second):
		// Update instance status
		mockMutex.Lock()
		if instance, exists := mockInstances[name]; exists {
			instance.Status = "running"
			instance.StartedAt = time.Now()
			instance.UpdatedAt = time.Now()
		}
		mockMutex.Unlock()

		return c.GetInstance(ctx, name)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
