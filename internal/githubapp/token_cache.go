package githubapp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenCache caches a single installation token with expiry.
type TokenCache struct {
	mu       sync.Mutex
	token    string
	expires  time.Time
	leeway   time.Duration
	fetching bool
}

// NewTokenCache builds a cache ensuring tokens are refreshed before expiry.
func NewTokenCache(leeway time.Duration) *TokenCache {
	return &TokenCache{
		leeway: leeway,
	}
}

// Get returns a cached token or fetches a new one using fetch.
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
