package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
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
