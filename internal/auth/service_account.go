package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func NewServiceAccountTokenSource(ctx context.Context, keyPath string, scopes []string) (oauth2.TokenSource, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key: %w", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, data, scopes...) //nolint:staticcheck // TODO: migrate to cloud.google.com/go/auth
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	return creds.TokenSource, nil
}

func NewServiceAccountOption(keyPath string, scopes []string) (option.ClientOption, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key: %w", err)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), data, scopes...) //nolint:staticcheck // TODO: migrate to cloud.google.com/go/auth
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	return option.WithCredentials(creds), nil
}
