package models

import "time"

type TargetType string

const (
	TargetTypeServer   TargetType = "server"
	TargetTypeWebsite  TargetType = "website"
	TargetTypeAPI      TargetType = "api"
	TargetTypeDatabase TargetType = "database"
)

type TargetStatus string

const (
	TargetStatusOnline  TargetStatus = "ONLINE"
	TargetStatusOffline TargetStatus = "OFFLINE"
	TargetStatusPending TargetStatus = "PENDING"
)

type AuthConfig struct {
	// For Server (SSH)
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`

	// For Website / API
	URL                string            `json:"url,omitempty"`
	Method             string            `json:"method,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	ExpectedStatusCode int               `json:"expected_status_code,omitempty"`

	// For Database
	DBEngine string `json:"db_engine,omitempty"` // postgres, mysql, redis, etc.
	DBName   string `json:"db_name,omitempty"`
}

type Target struct {
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`
	Name            string       `json:"name"`
	Type            TargetType   `json:"type"`
	Host            string       `json:"host"`
	Port            int          `json:"port"`
	AuthConfig      string       `json:"auth_config"` // JSON serialized AuthConfig
	PollingInterval int          `json:"polling_interval"`
	Status          TargetStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`

	// Parsed for convenience
	Config *AuthConfig `json:"config,omitempty"`
}

type CreateTargetRequest struct {
	Name            string      `json:"name" binding:"required"`
	Type            TargetType  `json:"type" binding:"required"`
	Host            string      `json:"host" binding:"required"`
	Port            int         `json:"port"`
	PollingInterval int         `json:"polling_interval"`
	Config          AuthConfig  `json:"config"`
}

type UpdateTargetRequest struct {
	Name            string      `json:"name"`
	Type            TargetType  `json:"type"`
	Host            string      `json:"host"`
	Port            int         `json:"port"`
	PollingInterval int         `json:"polling_interval"`
	Config          *AuthConfig `json:"config,omitempty"`
}

type TestConnectionRequest struct {
	Type   TargetType `json:"type" binding:"required"`
	Host   string     `json:"host" binding:"required"`
	Port   int        `json:"port"`
	Config AuthConfig `json:"config"`
}

type TestConnectionResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Latency float64           `json:"latency_ms"`
	Details map[string]any    `json:"details,omitempty"`
}
