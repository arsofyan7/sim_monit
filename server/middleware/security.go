package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders applies essential OWASP recommended HTTP security headers
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent Clickjacking (framing)
		c.Header("X-Frame-Options", "DENY")

		// Legacy XSS filter protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control Referrer header leakage
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict dangerous browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// Content Security Policy (allows Vue 3 and Vite in dev, restrictive in prod)
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: https:; connect-src 'self' http: https: ws: wss:; frame-ancestors 'none';"
		c.Header("Content-Security-Policy", csp)

		// HTTP Strict Transport Security in production mode
		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}

// BodySizeLimiter restricts incoming request payload size to prevent DoS via large bodies
func BodySizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
