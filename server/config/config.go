package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	StaticDir      string
	GinMode        string
	AllowedOrigins []string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./monitoring.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "sim_monit_default_super_secret_jwt_key_2026"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./frontend/dist"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://monitor.siberhub.id",
		}
	}

	return &Config{
		Port:           port,
		DBPath:         dbPath,
		JWTSecret:      jwtSecret,
		StaticDir:      staticDir,
		GinMode:        ginMode,
		AllowedOrigins: origins,
	}
}
