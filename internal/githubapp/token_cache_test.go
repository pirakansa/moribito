package githubapp

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTokenCacheReuse(t *testing.T) {
	cache := NewTokenCache(30 * time.Second)
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	calls := 0
	fetch := func(_ context.Context) (InstallationTokenResponse, error) {
		calls++
		return InstallationTokenResponse{
			Token:     "token-1",
			ExpiresAt: now.Add(2 * time.Hour),
		}, nil
	}

	token, err := cache.Get(context.Background(), now, fetch)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if token != "token-1" {
		t.Fatalf("unexpected token: %s", token)
	}

	token, err = cache.Get(context.Background(), now.Add(1*time.Minute), fetch)
	if err != nil {
		t.Fatalf("get cached: %v", err)
	}
	if token != "token-1" {
		t.Fatalf("unexpected token: %s", token)
	}
	if calls != 1 {
		t.Fatalf("expected 1 fetch call, got %d", calls)
	}
}

func TestTokenCacheRefreshesNearExpiry(t *testing.T) {
	cache := NewTokenCache(30 * time.Second)
	start := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	calls := 0
	fetch := func(_ context.Context) (InstallationTokenResponse, error) {
		calls++
		return InstallationTokenResponse{
			Token:     fmt.Sprintf("token-%d", calls),
			ExpiresAt: start.Add(1 * time.Minute),
		}, nil
	}

	_, err := cache.Get(context.Background(), start, fetch)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	_, err = cache.Get(context.Background(), start.Add(45*time.Second), fetch)
	if err != nil {
		t.Fatalf("get refresh: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 fetch calls, got %d", calls)
	}
}
