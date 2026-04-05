package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("GOOGLE_AUTH_TYPE", "service-account")
	os.Setenv("GOOGLE_SERVICE_ACCOUNT_PATH", "./test-key.json")
	defer os.Unsetenv("GOOGLE_AUTH_TYPE")
	defer os.Unsetenv("GOOGLE_SERVICE_ACCOUNT_PATH")

	cfg := Load()
	assert.Equal(t, 3000, cfg.App.Port)
	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, "service-account", cfg.Google.AuthType)
}

func TestLoad_InvalidEnv_Panics(t *testing.T) {
	os.Setenv("NODE_ENV", "staging")
	defer os.Unsetenv("NODE_ENV")

	assert.Panics(t, func() {
		Load()
	})
}

func TestLoad_OAuth2_MissingClientID_Panics(t *testing.T) {
	os.Setenv("GOOGLE_AUTH_TYPE", "oauth2")
	os.Setenv("GOOGLE_CLIENT_ID", "")
	os.Setenv("GOOGLE_CLIENT_SECRET", "real-secret")
	defer os.Unsetenv("GOOGLE_AUTH_TYPE")
	defer os.Unsetenv("GOOGLE_CLIENT_ID")
	defer os.Unsetenv("GOOGLE_CLIENT_SECRET")

	assert.Panics(t, func() {
		Load()
	})
}

func TestLoad_InvalidAuthType_Panics(t *testing.T) {
	os.Setenv("GOOGLE_AUTH_TYPE", "invalid")
	defer os.Unsetenv("GOOGLE_AUTH_TYPE")

	assert.Panics(t, func() {
		Load()
	})
}
