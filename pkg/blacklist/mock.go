package blacklist

import (
	"context"
	"sync"
	"time"
)

type InMemoryTokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token_hash -> expiry_time
}

func NewInMemoryTokenBlacklist() *InMemoryTokenBlacklist {
	return &InMemoryTokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

func (m *InMemoryTokenBlacklist) InsertToBlacklist(ctx context.Context, token string, expiresAt time.Time) error {
	hash := hashToken(token)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[hash] = expiresAt
	return nil
}

func (m *InMemoryTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	hash := hashToken(token)

	m.mu.RLock()
	expiry, exists := m.tokens[hash]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(expiry) {
		m.mu.Lock()
		delete(m.tokens, hash)
		m.mu.Unlock()

		return false, nil
	}

	return true, nil
}
