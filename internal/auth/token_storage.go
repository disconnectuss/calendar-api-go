package auth

import (
	"sync"
	"time"
)

type StoredTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiryDate   time.Time
	Email        string
}

type TokenStorage struct {
	mu     sync.RWMutex
	tokens map[string]*StoredTokens
}

func NewTokenStorage() *TokenStorage {
	return &TokenStorage{
		tokens: make(map[string]*StoredTokens),
	}
}

func (ts *TokenStorage) Store(sessionID string, tokens *StoredTokens) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tokens[sessionID] = tokens
}

func (ts *TokenStorage) Get(sessionID string) (*StoredTokens, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tokens[sessionID]
	return t, ok
}

func (ts *TokenStorage) Remove(sessionID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tokens, sessionID)
}

func (ts *TokenStorage) GetByAccessToken(accessToken string) (*StoredTokens, string, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	for sessionID, t := range ts.tokens {
		if t.AccessToken == accessToken {
			return t, sessionID, true
		}
	}
	return nil, "", false
}
