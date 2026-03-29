package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

type counterWindow struct {
	start time.Time
	count int
}

type fixedWindowLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	maxRequests int
	buckets     map[string]counterWindow
	lastCleanup time.Time
}

func NewRateLimitMiddleware(log *logger.Logger, name string, maxRequests int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	if maxRequests <= 0 || window <= 0 || keyFunc == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := &fixedWindowLimiter{
		window:      window,
		maxRequests: maxRequests,
		buckets:     make(map[string]counterWindow),
	}

	return func(c *gin.Context) {
		key := strings.TrimSpace(keyFunc(c))
		if key == "" {
			key = fmt.Sprintf("%s|%s|%s", c.Request.Method, c.FullPath(), RemoteIP(c))
		}

		allowed, retryAfter := limiter.allow(key, time.Now().UTC())
		if allowed {
			c.Next()
			return
		}

		retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
		if retryAfterSeconds < 1 {
			retryAfterSeconds = 1
		}

		c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
		log.InfoContext(c.Request.Context(), "request throttled", "limiter", name, "path", c.FullPath(), "method", c.Request.Method, "remote_ip", RemoteIP(c))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited", "message": "too many requests, retry later"})
	}
}

func AuthenticatedRouteKey(c *gin.Context) string {
	if account, ok := AuthenticatedAccount(c); ok {
		return fmt.Sprintf("%s|%s|user:%d", c.Request.Method, c.FullPath(), account.UserID)
	}

	return ClientRouteKey(c)
}

func ClientRouteKey(c *gin.Context) string {
	return fmt.Sprintf("%s|%s|ip:%s", c.Request.Method, c.FullPath(), RemoteIP(c))
}

func RemoteIP(c *gin.Context) string {
	remoteAddr := strings.TrimSpace(c.Request.RemoteAddr)
	if remoteAddr == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}

	return remoteAddr
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= l.window {
		for bucketKey, bucket := range l.buckets {
			if now.Sub(bucket.start) >= l.window {
				delete(l.buckets, bucketKey)
			}
		}
		l.lastCleanup = now
	}

	bucket, ok := l.buckets[key]
	if !ok || now.Sub(bucket.start) >= l.window {
		l.buckets[key] = counterWindow{start: now, count: 1}
		return true, 0
	}

	if bucket.count >= l.maxRequests {
		return false, l.window - now.Sub(bucket.start)
	}

	bucket.count++
	l.buckets[key] = bucket
	return true, 0
}
