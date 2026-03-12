package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter handles usage quotas using Redis.
type RateLimiter struct {
	rdb *redis.Client
}

// NewRateLimiter creates a new limiter instance.
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// Allow checks if a key (e.g., project_id:feature) is within its limit for a given period.
// It returns true if the action is allowed, false otherwise.
func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, period time.Duration) (bool, error) {
	if l.rdb == nil {
		// If Redis is not available, fail open or closed? 
		// For security, we'll fail closed for AI features.
		return false, fmt.Errorf("redis connection unavailable")
	}

	redisKey := fmt.Sprintf("limiter:%s", key)

	// Increment the counter
	count, err := l.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	// If this is the first hit, set expiration
	if count == 1 {
		l.rdb.Expire(ctx, redisKey, period)
	}

	// Check if limit exceeded
	if int(count) > limit {
		return false, nil
	}

	return true, nil
}

// GetRemaining returns how many requests are left for a key.
func (l *RateLimiter) GetRemaining(ctx context.Context, key string, limit int) (int, error) {
	redisKey := fmt.Sprintf("limiter:%s", key)
	val, err := l.rdb.Get(ctx, redisKey).Int()
	if err == redis.Nil {
		return limit, nil
	}
	if err != nil {
		return 0, err
	}
	remaining := limit - val
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}
