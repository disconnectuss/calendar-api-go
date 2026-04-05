package auth

import (
	"api-go/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewOAuth2Config(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURI,
		Scopes:       cfg.Google.Scopes,
		Endpoint:     google.Endpoint,
	}
}
