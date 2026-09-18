package routes

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sim_monit/server/config"
	"sim_monit/server/handlers"
	"sim_monit/server/middleware"
)

func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 1. Security Headers (Anti-XSS, Anti-Clickjacking, CSP, Referrer-Policy)
	r.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))

	// 2. Request Body Size Limiter (Max 1MB payload to prevent DoS)
	r.Use(middleware.BodySizeLimiter(1 << 20))

	// 3. Secure CORS Middleware using Configured Allowed Origins
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range cfg.AllowedOrigins {
		allowedOriginsMap[strings.ToLower(strings.TrimRight(origin, "/"))] = true
	}

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		normalizedOrigin := strings.ToLower(strings.TrimRight(origin, "/"))

		if origin != "" && allowedOriginsMap[normalizedOrigin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		if c.Request.Method == "OPTIONS" {
			if origin != "" && allowedOriginsMap[normalizedOrigin] {
				c.AbortWithStatus(http.StatusNoContent)
			} else {
				c.AbortWithStatus(http.StatusForbidden)
			}
			return
		}
		c.Next()
	})

	authHandler := handlers.NewAuthHandler(db, cfg)
	targetHandler := handlers.NewTargetHandler(db)
	metricHandler := handlers.NewMetricHandler(db)

	api := r.Group("/api")
	// Apply general API rate limiter: 120 req/min per IP
	api.Use(middleware.RateLimitMiddleware(120, time.Minute, "Terlalu banyak request ke API. Silakan coba lagi beberapa saat."))
	{
		// Public Auth routes with strict brute-force rate limiter: 10 req/min per IP
		authLimiter := middleware.RateLimitMiddleware(10, time.Minute, "Terlalu banyak percobaan autentikasi. Silakan tunggu 1 menit sebelum mencoba lagi.")
		auth := api.Group("/auth")
		{
			auth.POST("/register", authLimiter, authHandler.Register)
			auth.POST("/login", authLimiter, authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.Me)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Target routes
			targets := protected.Group("/targets")
			{
				targets.GET("", targetHandler.GetTargets)
				targets.POST("", targetHandler.CreateTarget)
				targets.GET("/:id", targetHandler.GetTarget)
				targets.PUT("/:id", targetHandler.UpdateTarget)
				targets.DELETE("/:id", targetHandler.DeleteTarget)
				targets.POST("/test-connection", targetHandler.TestConnection)

				// Metric routes under target
				targets.GET("/:id/metrics", metricHandler.GetTargetMetrics)
				targets.GET("/:id/stats", metricHandler.GetTargetStats)
			}
		}
	}

	// Serve Static Vue Frontend if static directory exists
	staticDir := cfg.StaticDir
	if _, err := os.Stat(staticDir); err == nil {
		r.Use(func(c *gin.Context) {
			path := c.Request.URL.Path

			// Don't intercept API requests
			if strings.HasPrefix(path, "/api") {
				c.Next()
				return
			}

			// Check if file physically exists in staticDir
			filePath := filepath.Join(staticDir, filepath.Clean(path))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				c.File(filePath)
				c.Abort()
				return
			}

			// SPA Fallback: Serve index.html
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
				c.Abort()
				return
			}

			c.Next()
		})
	}

	return r
}
