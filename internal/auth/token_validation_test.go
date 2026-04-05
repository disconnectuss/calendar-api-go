package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToken_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer valid-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"email":"user@example.com","email_verified":true,"name":"Test User","sub":"12345"}`))
	}))
	defer server.Close()

	original := userinfoEndpoint
	userinfoEndpoint = server.URL
	defer func() { userinfoEndpoint = original }()

	info, err := ValidateToken(context.Background(), "valid-token")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", info.Email)
	assert.Equal(t, "Test User", info.Name)
}

func TestValidateToken_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	original := userinfoEndpoint
	userinfoEndpoint = server.URL
	defer func() { userinfoEndpoint = original }()

	_, err := ValidateToken(context.Background(), "invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestValidateToken_NetworkError(t *testing.T) {
	original := userinfoEndpoint
	userinfoEndpoint = "http://localhost:1" // unreachable
	defer func() { userinfoEndpoint = original }()

	_, err := ValidateToken(context.Background(), "any-token")
	assert.Error(t, err)
}
