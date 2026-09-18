package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientVisitor struct {
	tokens     float64
	lastAccess time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*clientVisitor
	rate     float64 // tokens added per second
	burst    float64 // max token capacity
}

func NewIPRateLimiter(maxRequests int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors: make(map[string]*clientVisitor),
		rate:     float64(maxRequests) / window.Seconds(),
		burst:    float64(maxRequests),
	}

	// Background routine to clean up stale visitors every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, v := range limiter.visitors {
				if v.lastAccess.Before(cutoff) {
					delete(limiter.visitors, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	v, exists := l.visitors[ip]
	if !exists {
		l.visitors[ip] = &clientVisitor{
			tokens:     l.burst - 1,
			lastAccess: now,
		}
		return true
	}

	// Replenish tokens based on elapsed time
	elapsed := now.Sub(v.lastAccess).Seconds()
	v.lastAccess = now
	v.tokens += elapsed * l.rate
	if v.tokens > l.burst {
		v.tokens = l.burst
	}

	if v.tokens >= 1 {
		v.tokens -= 1
		return true
	}

	return false
}

// RateLimitMiddleware creates a Gin middleware for rate limiting by client IP
func RateLimitMiddleware(maxRequests int, window time.Duration, message string) gin.HandlerFunc {
	limiter := NewIPRateLimiter(maxRequests, window)

	return func(c *gin.Context) {
		ip := getClientIP(c)

		if !limiter.Allow(ip) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": message,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For first
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	return c.ClientIP()
}
