package blacklist

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlacklist interface {
	InsertToBlacklist(ctx context.Context, token string, expiresAt time.Time) error

	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

type RedisTokenBlacklist struct {
	client *redis.Client
}

func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *RedisTokenBlacklist) InsertToBlacklist(ctx context.Context, token string, expiresAt time.Time) error {
	hash := hashToken(token)
	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		return nil
	}

	key := fmt.Sprintf("blacklist:token:%s", hash)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

func (r *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	hash := hashToken(token)
	key := fmt.Sprintf("blacklist:token:%s", hash)

	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val > 0, nil
}
