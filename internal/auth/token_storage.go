package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const tokenTTL = 1 * time.Hour

type StoredTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiryDate   time.Time
	Email        string
	CreatedAt    time.Time
}

type TokenStorage struct {
	mu             sync.RWMutex
	tokens         map[string]*StoredTokens
	accessTokenIdx map[string]string // accessToken -> sessionID
	oauthStates    map[string]time.Time
}

func NewTokenStorage() *TokenStorage {
	ts := &TokenStorage{
		tokens:         make(map[string]*StoredTokens),
		accessTokenIdx: make(map[string]string),
		oauthStates:    make(map[string]time.Time),
	}
	go ts.cleanupLoop()
	return ts
}

func (ts *TokenStorage) Store(sessionID string, tokens *StoredTokens) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if old, ok := ts.tokens[sessionID]; ok {
		delete(ts.accessTokenIdx, old.AccessToken)
	}

	tokens.CreatedAt = time.Now()
	ts.tokens[sessionID] = tokens
	ts.accessTokenIdx[tokens.AccessToken] = sessionID
}

func (ts *TokenStorage) Get(sessionID string) (*StoredTokens, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tokens[sessionID]
	if !ok {
		return nil, false
	}
	if time.Since(t.CreatedAt) > tokenTTL {
		return nil, false
	}
	return t, ok
}

func (ts *TokenStorage) Remove(sessionID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tokens[sessionID]; ok {
		delete(ts.accessTokenIdx, t.AccessToken)
	}
	delete(ts.tokens, sessionID)
}

func (ts *TokenStorage) GetByAccessToken(accessToken string) (*StoredTokens, string, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	sessionID, ok := ts.accessTokenIdx[accessToken]
	if !ok {
		return nil, "", false
	}
	t := ts.tokens[sessionID]
	if time.Since(t.CreatedAt) > tokenTTL {
		return nil, "", false
	}
	return t, sessionID, true
}

func (ts *TokenStorage) GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	ts.mu.Lock()
	ts.oauthStates[state] = time.Now()
	ts.mu.Unlock()

	return state, nil
}

func (ts *TokenStorage) ValidateOAuthState(state string) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	created, ok := ts.oauthStates[state]
	if !ok {
		return false
	}
	delete(ts.oauthStates, state)
	return time.Since(created) < 10*time.Minute
}

func (ts *TokenStorage) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ts.mu.Lock()
		now := time.Now()
		for sid, t := range ts.tokens {
			if now.Sub(t.CreatedAt) > tokenTTL {
				delete(ts.accessTokenIdx, t.AccessToken)
				delete(ts.tokens, sid)
			}
		}
		for state, created := range ts.oauthStates {
			if now.Sub(created) > 10*time.Minute {
				delete(ts.oauthStates, state)
			}
		}
		ts.mu.Unlock()
	}
}
