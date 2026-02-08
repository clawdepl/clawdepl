// Package auth handles authentication flows for clawdpl.
package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/moltyverse/clawdpl/internal/config"
)

const (
	// AuthBaseURL is the base URL for authentication
	AuthBaseURL = "https://clawdpl.dev"
	// AuthPath is the path for CLI authentication
	AuthPath = "/cli/auth"
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

	// Build auth URL
	authURL := fmt.Sprintf("%s%s?callback=%s&state=%s", AuthBaseURL, AuthPath, callbackURL, state)

	fmt.Printf("Opening browser to authenticate...\n")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically.\n")
	}

	// Wait for callback
	credsChan := make(chan *config.Credentials, 1)
	errChan := make(chan error, 1)

	server := &http.Server{}
	http.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would validate the state and exchange the code
		// For now, we'll mock the response
		receivedState := r.URL.Query().Get("state")
		if receivedState != state {
			errChan <- fmt.Errorf("state mismatch")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		// Mock: In reality, we'd exchange the code for tokens here
		// For now, create mock credentials
		creds := &config.Credentials{
			AccessToken:  "mock_access_token_" + state[:8],
			RefreshToken: "mock_refresh_token_" + state[:8],
			User: &config.User{
				ID:    "user_" + state[:8],
				Email: "user@example.com",
				Name:  "Demo User",
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>clawdpl - Login Successful</title></head>
<body style="font-family: system-ui; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #1a1a2e;">
<div style="text-align: center; color: #eee;">
<h1>✓ Login Successful</h1>
<p>You can close this window and return to your terminal.</p>
</div>
</body>
</html>`)

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
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := fmt.Sprintf("%s%s?state=%s", AuthBaseURL, AuthPath, state)

	fmt.Printf("Visit this URL to authenticate:\n\n  %s\n\n", authURL)
	fmt.Printf("After authenticating, paste the token here: ")

	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, fmt.Errorf("no token provided")
	}

	// Mock: In reality, we'd validate the token with the server
	creds := &config.Credentials{
		AccessToken:  token,
		RefreshToken: "mock_refresh_" + token[:8],
		User: &config.User{
			ID:    "user_" + token[:8],
			Email: "user@example.com",
			Name:  "Demo User",
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return creds, nil
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
