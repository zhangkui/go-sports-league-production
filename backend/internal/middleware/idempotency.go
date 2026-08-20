package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/goxm2/sports-league/internal/pkg/response"
)

// IdempotencyMiddleware caches POST/PUT responses keyed by the Idempotency-Key
// header, ensuring INV-07 (idempotent key operations).
type IdempotencyMiddleware struct {
	RDB    *redis.Client
	TTL    time.Duration
}

func NewIdempotencyMiddleware(rdb *redis.Client, ttl time.Duration) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{RDB: rdb, TTL: ttl}
}

type cachedResponse struct {
	Status int
	Body   []byte
}

const (
	idemKeyPrefix = "idem:"
	idemLockSuffix = ":lock"
)

// Wrap enforces idempotency for handlers that opt in.
func (im *IdempotencyMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if im == nil || im.RDB == nil {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("Idempotency-Key")
		if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPut) {
			next.ServeHTTP(w, r)
			return
		}
		redisKey := idemKeyPrefix + key
		ctx := r.Context()

		// short body for replay
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		// try to claim processing
		ok, err := im.RDB.SetNX(ctx, redisKey+idemLockSuffix, "1", im.TTL).Result()
		if err == nil && !ok {
			// another request is processing / has processed; replay cached
			cached, e := im.load(ctx, redisKey)
			if e == nil && cached != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Idempotent-Replay", "true")
				w.WriteHeader(cached.Status)
				_, _ = w.Write(cached.Body)
				return
			}
			response.Error(w, r, http.StatusConflict, response.CodeConflict, "duplicate idempotent request in progress")
			return
		}

		rec := &recordingWriter{header: http.Header{}, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// cache only successful / 4xx
		if rec.status < 500 {
			_ = im.store(ctx, redisKey, cachedResponse{Status: rec.status, Body: rec.body.Bytes()})
		}
		// forward to real writer
		for k, vs := range rec.header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.status)
		_, _ = w.Write(rec.body.Bytes())
	})
}

func (im *IdempotencyMiddleware) store(ctx context.Context, key string, c cachedResponse) error {
	if im == nil || im.RDB == nil {
		return nil
	}
	pipe := im.RDB.TxPipeline()
	pipe.HSet(ctx, key, "status", c.Status, "body", c.Body)
	pipe.Expire(ctx, key, im.TTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (im *IdempotencyMiddleware) load(ctx context.Context, key string) (*cachedResponse, error) {
	if im == nil || im.RDB == nil {
		return nil, nil
	}
	status, err := im.RDB.HGet(ctx, key, "status").Int()
	if err != nil {
		return nil, err
	}
	body, err := im.RDB.HGet(ctx, key, "body").Bytes()
	if err != nil {
		return nil, err
	}
	return &cachedResponse{Status: status, Body: body}, nil
}

type recordingWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (rw *recordingWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = http.Header{}
	}
	return rw.header
}
func (rw *recordingWriter) WriteHeader(statusCode int) { rw.status = statusCode }
func (rw *recordingWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}
