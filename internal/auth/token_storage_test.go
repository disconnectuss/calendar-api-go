package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenStorage_Store_And_Get(t *testing.T) {
	ts := NewTokenStorage()

	tokens := &StoredTokens{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiryDate:   time.Now().Add(time.Hour),
		Email:        "test@example.com",
	}

	ts.Store("session-1", tokens)

	got, ok := ts.Get("session-1")
	require.True(t, ok)
	assert.Equal(t, "access-123", got.AccessToken)
	assert.Equal(t, "test@example.com", got.Email)
}

func TestTokenStorage_Get_NotFound(t *testing.T) {
	ts := NewTokenStorage()

	_, ok := ts.Get("nonexistent")
	assert.False(t, ok)
}

func TestTokenStorage_Remove(t *testing.T) {
	ts := NewTokenStorage()

	ts.Store("session-1", &StoredTokens{AccessToken: "token"})
	ts.Remove("session-1")

	_, ok := ts.Get("session-1")
	assert.False(t, ok)
}

func TestTokenStorage_Remove_Cleans_AccessTokenIndex(t *testing.T) {
	ts := NewTokenStorage()

	ts.Store("session-1", &StoredTokens{AccessToken: "token-abc"})
	ts.Remove("session-1")

	_, _, ok := ts.GetByAccessToken("token-abc")
	assert.False(t, ok)
}

func TestTokenStorage_GetByAccessToken(t *testing.T) {
	ts := NewTokenStorage()

	tokens := &StoredTokens{
		AccessToken: "unique-token",
		Email:       "user@example.com",
	}
	ts.Store("session-abc", tokens)

	got, sessionID, ok := ts.GetByAccessToken("unique-token")
	require.True(t, ok)
	assert.Equal(t, "session-abc", sessionID)
	assert.Equal(t, "user@example.com", got.Email)
}

func TestTokenStorage_GetByAccessToken_NotFound(t *testing.T) {
	ts := NewTokenStorage()

	_, _, ok := ts.GetByAccessToken("nonexistent")
	assert.False(t, ok)
}

func TestTokenStorage_Store_Replaces_Old_AccessToken_Index(t *testing.T) {
	ts := NewTokenStorage()

	ts.Store("session-1", &StoredTokens{AccessToken: "old-token"})
	ts.Store("session-1", &StoredTokens{AccessToken: "new-token"})

	_, _, ok := ts.GetByAccessToken("old-token")
	assert.False(t, ok, "old access token should be removed from index")

	_, sid, ok := ts.GetByAccessToken("new-token")
	require.True(t, ok)
	assert.Equal(t, "session-1", sid)
}

func TestTokenStorage_OAuthState(t *testing.T) {
	ts := NewTokenStorage()

	state, err := ts.GenerateOAuthState()
	require.NoError(t, err)
	assert.Len(t, state, 64) // 32 bytes hex encoded

	assert.True(t, ts.ValidateOAuthState(state), "state should be valid on first use")
	assert.False(t, ts.ValidateOAuthState(state), "state should be consumed after first use")
}

func TestTokenStorage_OAuthState_Invalid(t *testing.T) {
	ts := NewTokenStorage()
	assert.False(t, ts.ValidateOAuthState("bogus"))
}
