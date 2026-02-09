// Package auth handles authentication flows for clawdepl.
package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/clawdepl/clawdepl/internal/config"
)

const (
	// AuthBaseURL is the base URL for authentication
	AuthBaseURL = "https://clawdepl.dev"
	// AuthPath is the path for CLI authentication
	AuthPath = "/auth/cli"
	// CallbackPath is the path for OAuth callback
	CallbackPath = "/callback"
)

// generateState generates a random state string for OAuth
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// openBrowser opens the default browser to the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// ValidateToken validates a token against the backend and returns user info
func ValidateToken(ctx context.Context, token string) (*config.Credentials, error) {
	// Call Convex /api/users/verify with Bearer token
	verifyURL := fmt.Sprintf("%s/api/users/verify", buildinfo.ConvexEndpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token validation failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("token validation failed (status %d)", resp.StatusCode)
	}

	// Parse successful response
	var verifyResp struct {
		User struct {
			UserID   string `json:"userId"`
			Username string `json:"username"`
			Type     string `json:"type"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	creds := &config.Credentials{
		AccessToken: token,
		User: &config.User{
			ID:    verifyResp.User.UserID,
			Email: verifyResp.User.Username, // Username might be email
			Name:  verifyResp.User.Username,
		},
	}

	return creds, nil
}

// LoginWithToken performs login using a pre-obtained token
func LoginWithToken(ctx context.Context, token string) (*config.Credentials, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	// Validate token and get user info
	creds, err := ValidateToken(ctx, token)
	if err != nil {
		// If validation fails, still allow storing the token
		// (the backend might not have the verify endpoint yet)
		fmt.Printf("Warning: Could not validate token: %v\n", err)
		fmt.Printf("Storing token anyway. It will be validated on first API call.\n")

		creds = &config.Credentials{
			AccessToken: token,
			User: &config.User{
				ID:    "unknown",
				Email: "unknown",
				Name:  "Unknown User",
			},
		}
	}

	// Save credentials
	if err := config.SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	return creds, nil
}

// LoginWithBrowser performs OAuth login by opening a browser
func LoginWithBrowser(ctx context.Context) (*config.Credentials, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, CallbackPath)

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build auth URL with URL-encoded callback
	authURL := fmt.Sprintf("%s%s?callback=%s&state=%s", AuthBaseURL, AuthPath, url.QueryEscape(callbackURL), state)

	fmt.Printf("Opening browser to authenticate...\n")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically.\n")
	}

	// Wait for callback
	credsChan := make(chan *config.Credentials, 1)
	errChan := make(chan error, 1)

	// Create a new ServeMux to avoid conflicts with default mux
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Validate state to prevent CSRF
		receivedState := query.Get("state")
		if receivedState != state {
			errChan <- fmt.Errorf("state mismatch: expected %s, got %s", state, receivedState)
			http.Error(w, "Invalid state - possible CSRF attack", http.StatusBadRequest)
			return
		}

		// Parse token and user info from query params
		token := query.Get("token")
		if token == "" {
			errChan <- fmt.Errorf("no token received in callback")
			http.Error(w, "No token received", http.StatusBadRequest)
			return
		}

		userId := query.Get("userId")
		email := query.Get("email")
		name := query.Get("name")

		// Use email as fallback for name if not provided
		if name == "" {
			name = email
		}
		if name == "" {
			name = "User"
		}

		creds := &config.Credentials{
			AccessToken: token,
			User: &config.User{
				ID:    userId,
				Email: email,
				Name:  name,
			},
		}

		// Serve success page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>clawdepl - Login Successful</title></head>
<body style="font-family: system-ui; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #1a1a2e;">
<div style="text-align: center; color: #eee;">
<h1 style="color: #4ade80;">✓ Login Successful</h1>
<p>You can close this window and return to your terminal.</p>
<p style="color: #888; font-size: 0.9em;">Logged in as %s</p>
</div>
</body>
</html>`, email)

		credsChan <- creds
	})

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for result or timeout
	select {
	case creds := <-credsChan:
		server.Shutdown(ctx)
		return creds, nil
	case err := <-errChan:
		server.Shutdown(ctx)
		return nil, err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return nil, fmt.Errorf("login timed out")
	}
}

// LoginWithoutBrowser performs login without opening a browser
func LoginWithoutBrowser(ctx context.Context) (*config.Credentials, error) {
	authURL := fmt.Sprintf("%s%s", AuthBaseURL, AuthPath)

	fmt.Printf("Visit this URL to authenticate:\n\n  %s\n\n", authURL)
	fmt.Printf("After authenticating, copy the token and paste it here: ")

	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	// Use LoginWithToken to validate and store
	return LoginWithToken(ctx, token)
}

// LoginWithAPIKey performs login using an API key
func LoginWithAPIKey(ctx context.Context) (*config.Credentials, error) {
	fmt.Printf("Enter your API key: ")

	reader := bufio.NewReader(os.Stdin)
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	// Mock: In reality, we'd validate the API key with the server
	// and fetch user info
	creds := &config.Credentials{
		APIKey: apiKey,
		User: &config.User{
			ID:    "user_api_" + apiKey[:8],
			Email: "api-user@example.com",
			Name:  "API User",
		},
	}

	return creds, nil
}

// Login performs the login flow based on options
func Login(ctx context.Context, noBrowser bool, useAPIKey bool) (*config.Credentials, error) {
	var creds *config.Credentials
	var err error

	if useAPIKey {
		creds, err = LoginWithAPIKey(ctx)
	} else if noBrowser {
		creds, err = LoginWithoutBrowser(ctx)
	} else {
		creds, err = LoginWithBrowser(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Save credentials
	if err := config.SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("failed to save credentials: %w", err)
	}

	return creds, nil
}

// Logout clears the stored credentials
func Logout() error {
	return config.ClearCredentials()
}
