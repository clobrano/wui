package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// GetOAuth2Client creates an authenticated OAuth2 client for Google Calendar API
func GetOAuth2Client(ctx context.Context, credentialsPath, tokenPath string) (*calendar.Service, error) {
	// Read credentials file
	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// Parse the credentials
	config, err := google.ConfigFromJSON(credentials, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %w", err)
	}

	// Get token from file or request new one
	token, err := getTokenFromFile(tokenPath)
	if err != nil {
		slog.Info("No valid token found, requesting new authorization")
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, fmt.Errorf("unable to get authorization: %w", err)
		}
		if err := saveToken(tokenPath, token); err != nil {
			slog.Warn("Unable to save token", "error", err)
		}
	}

	// Create calendar service
	client := config.Client(ctx, token)
	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return srv, nil
}

// getTokenFromFile retrieves a token from a local file
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

// getTokenFromWeb uses Config to request a token from the web
// It starts a local HTTP server to receive the OAuth callback
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	// Use a local server to receive the callback
	// This is more user-friendly than copy-pasting codes
	codeCh := make(chan string)
	errCh := make(chan error)

	// Create a local HTTP server to handle the OAuth callback
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			fmt.Fprintf(w, "Error: No authorization code received")
			return
		}

		// Send success message to browser
		fmt.Fprintf(w, `
			<html>
			<head><title>Authorization Successful</title></head>
			<body style="font-family: sans-serif; text-align: center; padding: 50px;">
				<h1>✓ Authorization Successful</h1>
				<p>You can close this window and return to the terminal.</p>
			</body>
			</html>
		`)

		// Send the code through the channel
		codeCh <- code
	})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("failed to start local server: %w", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Update config to use localhost redirect
	config.RedirectURL = "http://localhost:8080"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n=== Google Calendar Authorization ===\n")
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open automatically, visit this URL:\n\n%s\n\n", authURL)

	// Try to open the browser automatically
	openBrowser(authURL)

	// Wait for either the code or an error
	var authCode string
	select {
	case authCode = <-codeCh:
		// Success - received the code
	case err := <-errCh:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("timeout waiting for authorization")
	}

	// Shutdown the server
	server.Shutdown(context.Background())

	// Exchange the code for a token
	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	fmt.Printf("✓ Authorization successful!\n\n")
	return token, nil
}

// openBrowser attempts to open the URL in the default browser
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

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) error {
	slog.Info("Saving credential file", "path", path)

	// Create directory if it doesn't exist
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
