package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// ErrNoToken is returned when no OAuth2 token file exists or can be loaded.
var ErrNoToken = errors.New("no OAuth2 token found")

// GetOAuth2Client creates an authenticated calendar service using the saved token file.
// If no token file exists, it returns ErrNoToken — callers must handle this by
// starting the OAuth2 authorization flow separately (see StartAuthServer).
func GetOAuth2Client(ctx context.Context, credentialsPath, tokenPath string) (*calendar.Service, error) {
	oauthConfig, err := LoadOAuthConfig(credentialsPath)
	if err != nil {
		return nil, err
	}

	token, err := getTokenFromFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoToken, err)
	}

	client := oauthConfig.Client(ctx, token)
	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return srv, nil
}

// LoadOAuthConfig parses the credentials file and returns an OAuth2 config.
func LoadOAuthConfig(credentialsPath string) (*oauth2.Config, error) {
	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credentials, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %w", err)
	}

	return oauthConfig, nil
}

// AuthServer manages the local HTTP server that receives the OAuth2 callback.
type AuthServer struct {
	// URL is the authorization URL the user must visit in their browser.
	URL    string
	codeCh chan string
	server *http.Server
	config *oauth2.Config
	ctx    context.Context
	cancel context.CancelFunc
}

// StartAuthServer starts a local callback server on a random port and returns
// an AuthServer whose URL field contains the authorization URL to show the user.
// The browser is NOT opened automatically — call OpenBrowser(authServer.URL) separately.
func StartAuthServer(oauthConfig *oauth2.Config) (*AuthServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	oauthConfig.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	codeCh := make(chan string, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Fprintf(w, "Error: no authorization code received")
			return
		}
		fmt.Fprintf(w, `<html><head><title>Authorized</title></head>`+
			`<body style="font-family:sans-serif;text-align:center;padding:50px">`+
			`<h1>&#10003; Authorization Successful</h1>`+
			`<p>You can close this window and return to the terminal.</p>`+
			`</body></html>`)
		select {
		case codeCh <- code:
		default:
		}
	})

	go func() {
		// Serve returns ErrServerClosed when Shutdown is called; ignore that.
		_ = server.Serve(listener)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	return &AuthServer{
		URL:    authURL,
		codeCh: codeCh,
		server: server,
		config: oauthConfig,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// WaitForToken blocks until the user completes authorization, Cancel is called,
// or the 5-minute timeout expires.
func (a *AuthServer) WaitForToken() (*oauth2.Token, error) {
	defer a.cancel()
	defer a.server.Shutdown(context.Background()) //nolint:errcheck

	select {
	case code := <-a.codeCh:
		token, err := a.config.Exchange(context.Background(), code)
		if err != nil {
			return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
		}
		return token, nil
	case <-a.ctx.Done():
		return nil, fmt.Errorf("authorization cancelled")
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timeout: no authorization received within 5 minutes")
	}
}

// Cancel cancels the authorization flow and shuts down the callback server.
// Safe to call multiple times.
func (a *AuthServer) Cancel() {
	a.cancel()
	a.server.Shutdown(context.Background()) //nolint:errcheck
}

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) {
	if err := openBrowser(url); err != nil {
		slog.Warn("Failed to open browser automatically", "url", url, "error", err)
	}
}

// SaveToken saves the given token to path (exported wrapper for saveToken).
func SaveToken(path string, token *oauth2.Token) error {
	return saveToken(path, token)
}

// IsTokenExpiredError returns true if the error indicates that the OAuth2 token
// has expired or been revoked and cannot be refreshed automatically.
func IsTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid_grant") ||
		strings.Contains(msg, "oauth2: token expired") ||
		strings.Contains(msg, "token has been expired") ||
		strings.Contains(msg, "Token has been expired")
}

// DeleteToken removes the token file from disk.
func DeleteToken(tokenPath string) error {
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to delete token file: %w", err)
	}
	return nil
}

// getTokenFromFile retrieves a token from a local file.
func getTokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// openBrowser attempts to open the URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	slog.Info("Saving credential file", "path", path)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create token directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}
