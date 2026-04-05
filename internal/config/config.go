package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig
	Google GoogleConfig
}

type AppConfig struct {
	Port           int      `mapstructure:"PORT"`
	Env            string   `mapstructure:"NODE_ENV"`
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}

type GoogleConfig struct {
	AuthType           string `mapstructure:"GOOGLE_AUTH_TYPE"`
	ClientID           string `mapstructure:"GOOGLE_CLIENT_ID"`
	ClientSecret       string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	RedirectURI        string `mapstructure:"GOOGLE_REDIRECT_URI"`
	Scopes             []string
	ServiceAccountPath string `mapstructure:"GOOGLE_SERVICE_ACCOUNT_PATH"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", 3000)
	viper.SetDefault("NODE_ENV", "development")
	viper.SetDefault("GOOGLE_AUTH_TYPE", "service-account")
	viper.SetDefault("GOOGLE_SERVICE_ACCOUNT_PATH", "./config/service-account-key.json")

	_ = viper.ReadInConfig() // .env dosyası yoksa sorun değil

	cfg := &Config{}

	cfg.App.Port = viper.GetInt("PORT")
	cfg.App.Env = viper.GetString("NODE_ENV")

	originsStr := viper.GetString("ALLOWED_ORIGINS")
	if originsStr != "" {
		cfg.App.AllowedOrigins = strings.Split(originsStr, ",")
	}

	validEnvs := map[string]bool{"development": true, "production": true, "test": true}
	if !validEnvs[cfg.App.Env] {
		panic(fmt.Sprintf("invalid NODE_ENV: %s (must be development, production, or test)", cfg.App.Env))
	}

	cfg.Google.AuthType = viper.GetString("GOOGLE_AUTH_TYPE")
	cfg.Google.ClientID = viper.GetString("GOOGLE_CLIENT_ID")
	cfg.Google.ClientSecret = viper.GetString("GOOGLE_CLIENT_SECRET")
	cfg.Google.RedirectURI = viper.GetString("GOOGLE_REDIRECT_URI")
	cfg.Google.ServiceAccountPath = viper.GetString("GOOGLE_SERVICE_ACCOUNT_PATH")
	cfg.Google.Scopes = []string{
		"https://www.googleapis.com/auth/calendar",
		"https://www.googleapis.com/auth/tasks",
	}

	switch cfg.Google.AuthType {
	case "oauth2":
		if cfg.Google.ClientID == "" || cfg.Google.ClientID == "your-client-id" {
			panic("GOOGLE_CLIENT_ID is required when GOOGLE_AUTH_TYPE=oauth2")
		}
		if cfg.Google.ClientSecret == "" || cfg.Google.ClientSecret == "your-client-secret" {
			panic("GOOGLE_CLIENT_SECRET is required when GOOGLE_AUTH_TYPE=oauth2")
		}
	case "service-account":
		if cfg.Google.ServiceAccountPath == "" {
			panic("GOOGLE_SERVICE_ACCOUNT_PATH is required when GOOGLE_AUTH_TYPE=service-account")
		}
	default:
		panic(fmt.Sprintf("invalid GOOGLE_AUTH_TYPE: %s (must be oauth2 or service-account)", cfg.Google.AuthType))
	}

	return cfg
}
