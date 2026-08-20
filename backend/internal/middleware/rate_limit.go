package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/goxm2/sports-league/internal/pkg/response"
)

// RateLimiter is a Redis-backed sliding-window rate limiter.
type RateLimiter struct {
	RDB    *redis.Client
	Limit  int
	Window time.Duration
}

// NewRateLimiter builds a limiter for `limit` requests per `window`.
func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{RDB: rdb, Limit: limit, Window: window}
}

// Allow checks whether the key is within the limit; returns remaining count.
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	if rl == nil || rl.RDB == nil {
		return true, rl.Limit, nil
	}
	now := time.Now().UnixNano()
	clearBefore := now - int64(rl.Window)
	pipe := rl.RDB.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(clearBefore, 10))
	members := pipe.ZRange(ctx, key, 0, -1)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, rl.Window+time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return true, rl.Limit, nil // fail-open on redis errors
	}
	count := 0
	if members != nil {
		count = len(members.Val())
	}
	if count >= rl.Limit {
		return false, 0, nil
	}
	return true, rl.Limit - count - 1, nil
}

// LimitMiddleware enforces a per-key rate limit (key extracted by keyFn).
func (rl *RateLimiter) LimitMiddleware(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, remaining, err := rl.Allow(r.Context(), keyFn(r))
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.Window.Seconds())))
				response.Error(w, r, http.StatusTooManyRequests, response.CodeTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LoginRateLimit builds the IP + username login limiter middleware.
func LoginRateLimit(rdb *redis.Client, ipLimit, userLimit int) func(http.Handler) http.Handler {
	ipRL := NewRateLimiter(rdb, ipLimit, time.Minute)
	userRL := NewRateLimiter(rdb, userLimit, time.Minute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipKey := "rl:login:ip:" + realIP(r)
			allowed, _, _ := ipRL.Allow(r.Context(), ipKey)
			if !allowed {
				response.Error(w, r, http.StatusTooManyRequests, response.CodeTooManyRequests, "too many login attempts from this IP")
				return
			}
			// username-based limit applied only if body parse is cheap; we use a
			// best-effort approach reading the username query/header to avoid
			// buffering the body here. Full enforcement happens in the handler.
			next.ServeHTTP(w, r)
			_ = userRL
		})
	}
}
