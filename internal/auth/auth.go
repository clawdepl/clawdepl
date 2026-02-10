// Package auth handles authentication flows for clawdepl.
package auth

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
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

var validateHTTPClient = &http.Client{Timeout: 30 * time.Second}

//go:embed callback_success.html
var callbackSuccessHTML string

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
	// Call POST /verify-token on the Supabase API
	verifyURL := fmt.Sprintf("%s/verify-token", buildinfo.APIEndpoint)

	tokenBody := map[string]string{"token": token}
	jsonBody, err := json.Marshal(tokenBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := validateHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Valid bool   `json:"valid"`
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token validation failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("token validation failed (status %d)", resp.StatusCode)
	}

	// Parse successful response: {valid, userId, email, name}
	var verifyResp struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"userId"`
		Email  string `json:"email"`
		Name   string `json:"name"`
	}

	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !verifyResp.Valid {
		return nil, fmt.Errorf("token validation failed: token is not valid")
	}

	// Use email as fallback for name
	name := verifyResp.Name
	if name == "" {
		name = verifyResp.Email
	}

	creds := &config.Credentials{
		AccessToken: token,
		User: &config.User{
			ID:    verifyResp.UserID,
			Email: verifyResp.Email,
			Name:  name,
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
		return nil, fmt.Errorf("token validation failed: %w", err)
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
	stdinChan := make(chan string, 1)

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

		creds, err := ValidateToken(r.Context(), token)
		if err != nil {
			errChan <- fmt.Errorf("token validation failed: %w", err)
			http.Error(w, "Token validation failed", http.StatusUnauthorized)
			return
		}

		if creds.User == nil {
			// Fallback to callback query params if user payload is unexpectedly absent.
			userID := query.Get("userId")
			email := query.Get("email")
			name := query.Get("name")
			if name == "" {
				name = email
			}
			if name == "" {
				name = "User"
			}
			creds.User = &config.User{
				ID:    userID,
				Email: email,
				Name:  name,
			}
		}

		// Serve styled success page (HTML reads query params via JS)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(callbackSuccessHTML))

		credsChan <- creds
	})

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Also listen for token paste from stdin (fallback if redirect doesn't work)
	fmt.Printf("Or paste your token here and press Enter: ")
	go func() {
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err == nil {
			token = strings.TrimSpace(token)
			if token != "" {
				stdinChan <- token
			}
		}
	}()

	// Wait for result or timeout
	select {
	case creds := <-credsChan:
		server.Shutdown(ctx)
		return creds, nil
	case token := <-stdinChan:
		server.Shutdown(ctx)
		// Validate and return credentials using the pasted token
		return LoginWithToken(ctx, token)
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

	return loginWithAPIKeyValue(ctx, apiKey)
}

func loginWithAPIKeyValue(ctx context.Context, apiKey string) (*config.Credentials, error) {
	// API keys are bearer credentials in the same auth system.
	validated, err := ValidateToken(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("API key validation failed: %w", err)
	}

	creds := &config.Credentials{
		APIKey: apiKey,
		User:   validated.User,
	}

	if creds.User == nil {
		creds.User = &config.User{
			ID:    deriveAPIKeyUserID(apiKey),
			Email: "unknown",
			Name:  "API User",
		}
	}

	return creds, nil
}

func deriveAPIKeyUserID(apiKey string) string {
	prefixLen := 8
	if len(apiKey) < prefixLen {
		prefixLen = len(apiKey)
	}
	return "user_api_" + apiKey[:prefixLen]
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
