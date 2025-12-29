package githubapp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenCache provides thread-safe caching for installation tokens.
// It automatically refreshes tokens before they expire.
type TokenCache struct {
	mu       sync.Mutex
	token    string
	expires  time.Time
	leeway   time.Duration
	fetching bool
}

// NewTokenCache creates a cache that refreshes tokens `leeway` before expiry.
// Recommended leeway: 1-2 minutes.
func NewTokenCache(leeway time.Duration) *TokenCache {
	return &TokenCache{
		leeway: leeway,
	}
}

// Get returns a cached token or fetches a new one using the provided function.
// It prevents concurrent fetches by returning an error if a fetch is in progress.
func (c *TokenCache) Get(ctx context.Context, now time.Time, fetch func(context.Context) (InstallationTokenResponse, error)) (string, error) {
	c.mu.Lock()
	if c.token != "" && now.Before(c.expires.Add(-c.leeway)) {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	if c.fetching {
		c.mu.Unlock()
		return "", fmt.Errorf("token fetch already in progress")
	}
	c.fetching = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.fetching = false
		c.mu.Unlock()
	}()

	resp, err := fetch(ctx)
	if err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", fmt.Errorf("token response missing token")
	}

	c.mu.Lock()
	c.token = resp.Token
	c.expires = resp.ExpiresAt
	c.mu.Unlock()
	return resp.Token, nil
}
