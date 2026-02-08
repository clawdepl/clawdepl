// Package config handles credential storage and configuration for clawdpl.
// Credentials are stored in ~/.clawdpl/credentials.json
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	// ConfigDirName is the name of the config directory in the user's home
	ConfigDirName = ".clawdpl"
	// CredentialsFileName is the name of the credentials file
	CredentialsFileName = "credentials.json"
)

// User represents the authenticated user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Credentials represents the stored authentication credentials
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	APIKey       string    `json:"api_key,omitempty"`
	User         *User     `json:"user,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// IsExpired checks if the credentials have expired
func (c *Credentials) IsExpired() bool {
	if c.APIKey != "" {
		return false // API keys don't expire
	}
	if c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// IsValid checks if the credentials are valid (non-empty and not expired)
func (c *Credentials) IsValid() bool {
	if c == nil {
		return false
	}
	if c.APIKey != "" {
		return true
	}
	return c.AccessToken != "" && !c.IsExpired()
}

// GetConfigDir returns the path to the clawdpl config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetCredentialsPath returns the path to the credentials file
func GetCredentialsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, CredentialsFileName), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(configDir, 0700)
}

// LoadCredentials loads credentials from the config file
func LoadCredentials() (*Credentials, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No credentials file, not an error
		}
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// SaveCredentials saves credentials to the config file
func SaveCredentials(creds *Credentials) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ClearCredentials removes the credentials file
func ClearCredentials() error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil // Already cleared
	}
	return err
}

// IsLoggedIn checks if the user is currently logged in with valid credentials
func IsLoggedIn() bool {
	creds, err := LoadCredentials()
	if err != nil {
		return false
	}
	return creds.IsValid()
}

// GetCurrentUser returns the current user if logged in
func GetCurrentUser() (*User, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	if creds == nil || !creds.IsValid() {
		return nil, nil
	}
	return creds.User, nil
}
