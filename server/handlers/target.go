package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sim_monit/server/collector"
	"sim_monit/server/models"
	"sim_monit/server/worker"
)

type TargetHandler struct {
	db *sql.DB
}

func NewTargetHandler(db *sql.DB) *TargetHandler {
	return &TargetHandler{db: db}
}

func maskSensitiveConfig(cfg *models.AuthConfig) {
	if cfg == nil {
		return
	}
	if cfg.Password != "" {
		cfg.Password = "********"
	}
	if cfg.PrivateKey != "" {
		cfg.PrivateKey = "********"
	}
}

func validateTarget(name string, tType models.TargetType, host string, port int, interval int) error {
	trimmedName := strings.TrimSpace(name)
	if len(trimmedName) < 2 || len(trimmedName) > 100 {
		return fmt.Errorf("nama target harus antara 2 hingga 100 karakter")
	}

	validTypes := map[models.TargetType]bool{
		models.TargetTypeServer:   true,
		models.TargetTypeWebsite:  true,
		models.TargetTypeAPI:      true,
		models.TargetTypeDatabase: true,
	}
	if !validTypes[tType] {
		return fmt.Errorf("tipe target tidak valid. Pilihan: server, website, api, database")
	}

	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" || len(trimmedHost) > 255 {
		return fmt.Errorf("host/URL tidak boleh kosong dan maksimal 255 karakter")
	}

	// Prevent control characters / newlines in host
	if strings.ContainsAny(trimmedHost, "\r\n\x00") {
		return fmt.Errorf("host mengandung karakter tidak valid")
	}

	// For website and API, check for disallowed URL schemes
	lowerHost := strings.ToLower(trimmedHost)
	if strings.HasPrefix(lowerHost, "file:") || strings.HasPrefix(lowerHost, "ftp:") ||
		strings.HasPrefix(lowerHost, "gopher:") || strings.HasPrefix(lowerHost, "javascript:") {
		return fmt.Errorf("skema protokol URL tidak diizinkan")
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("port harus antara 1 dan 65535")
	}

	if interval < 5 || interval > 86400 {
		return fmt.Errorf("polling interval harus antara 5 detik hingga 24 jam")
	}

	return nil
}

func (h *TargetHandler) GetTargets(c *gin.Context) {
	userID := c.GetInt64("user_id")

	rows, err := h.db.Query(`
		SELECT id, user_id, name, type, host, port, auth_config, polling_interval, status, created_at
		FROM targets
		WHERE user_id = ?
		ORDER BY id DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar target"})
		return
	}
	defer rows.Close()

	targets := make([]models.Target, 0)
	for rows.Next() {
		var t models.Target
		var authConfigStr string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Type, &t.Host, &t.Port, &authConfigStr, &t.PollingInterval, &t.Status, &t.CreatedAt); err != nil {
			continue
		}

		var cfg models.AuthConfig
		if err := json.Unmarshal([]byte(authConfigStr), &cfg); err == nil {
			// Mask sensitive credentials before returning to client
			maskSensitiveConfig(&cfg)
			t.Config = &cfg
			maskedJSON, _ := json.Marshal(cfg)
			t.AuthConfig = string(maskedJSON)
		} else {
			t.AuthConfig = "{}"
		}

		targets = append(targets, t)
	}

	c.JSON(http.StatusOK, targets)
}

func (h *TargetHandler) GetTarget(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target tidak valid"})
		return
	}

	var t models.Target
	var authConfigStr string
	err = h.db.QueryRow(`
		SELECT id, user_id, name, type, host, port, auth_config, polling_interval, status, created_at
		FROM targets
		WHERE id = ? AND user_id = ?
	`, targetID, userID).Scan(&t.ID, &t.UserID, &t.Name, &t.Type, &t.Host, &t.Port, &authConfigStr, &t.PollingInterval, &t.Status, &t.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target tidak ditemukan"})
		return
	}

	var cfg models.AuthConfig
	if err := json.Unmarshal([]byte(authConfigStr), &cfg); err == nil {
		maskSensitiveConfig(&cfg)
		t.Config = &cfg
		maskedJSON, _ := json.Marshal(cfg)
		t.AuthConfig = string(maskedJSON)
	} else {
		t.AuthConfig = "{}"
	}

	c.JSON(http.StatusOK, t)
}

func (h *TargetHandler) CreateTarget(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req models.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data input target tidak lengkap atau tidak valid"})
		return
	}

	if req.PollingInterval <= 0 {
		req.PollingInterval = 60
	}

	cleanName := strings.TrimSpace(req.Name)
	cleanHost := strings.TrimSpace(req.Host)

	if err := validateTarget(cleanName, req.Type, cleanHost, req.Port, req.PollingInterval); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authConfigJSON, err := json.Marshal(req.Config)
	if err != nil {
		authConfigJSON = []byte("{}")
	}

	res, err := h.db.Exec(`
		INSERT INTO targets (user_id, name, type, host, port, auth_config, polling_interval, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'PENDING', ?)
	`, userID, cleanName, string(req.Type), cleanHost, req.Port, string(authConfigJSON), req.PollingInterval, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan target baru"})
		return
	}

	targetID, _ := res.LastInsertId()

	// Internal copy with real credentials for scheduler
	internalTarget := models.Target{
		ID:              targetID,
		UserID:          userID,
		Name:            cleanName,
		Type:            req.Type,
		Host:            cleanHost,
		Port:            req.Port,
		AuthConfig:      string(authConfigJSON),
		PollingInterval: req.PollingInterval,
		Status:          models.TargetStatusPending,
		CreatedAt:       time.Now(),
		Config:          &req.Config,
	}

	// Trigger worker scheduler to start polling immediately with actual credentials
	if worker.GlobalScheduler != nil {
		worker.GlobalScheduler.ReloadTargets()
		worker.GlobalScheduler.TriggerImmediatePoll(&internalTarget)
	}

	// Client response copy with masked credentials
	maskedConfig := req.Config
	maskSensitiveConfig(&maskedConfig)
	maskedJSON, _ := json.Marshal(maskedConfig)

	clientResponse := internalTarget
	clientResponse.Config = &maskedConfig
	clientResponse.AuthConfig = string(maskedJSON)

	c.JSON(http.StatusCreated, clientResponse)
}

func (h *TargetHandler) UpdateTarget(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target tidak valid"})
		return
	}

	var req models.UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input update tidak valid"})
		return
	}

	// Verify target ownership
	var existing models.Target
	var authConfigStr string
	err = h.db.QueryRow(`
		SELECT id, user_id, name, type, host, port, auth_config, polling_interval, status, created_at
		FROM targets
		WHERE id = ? AND user_id = ?
	`, targetID, userID).Scan(&existing.ID, &existing.UserID, &existing.Name, &existing.Type, &existing.Host, &existing.Port, &authConfigStr, &existing.PollingInterval, &existing.Status, &existing.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target tidak ditemukan"})
		return
	}

	var existingCfg models.AuthConfig
	_ = json.Unmarshal([]byte(authConfigStr), &existingCfg)
	existing.Config = &existingCfg

	if req.Name != "" {
		existing.Name = strings.TrimSpace(req.Name)
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Host != "" {
		existing.Host = strings.TrimSpace(req.Host)
	}
	if req.Port > 0 {
		existing.Port = req.Port
	}
	if req.PollingInterval > 0 {
		existing.PollingInterval = req.PollingInterval
	}

	if err := validateTarget(existing.Name, existing.Type, existing.Host, existing.Port, existing.PollingInterval); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Config != nil {
		// If client sent masked placeholder, retain existing saved password/key
		if req.Config.Password == "********" {
			req.Config.Password = existingCfg.Password
		}
		if req.Config.PrivateKey == "********" {
			req.Config.PrivateKey = existingCfg.PrivateKey
		}

		cfgBytes, _ := json.Marshal(req.Config)
		existing.AuthConfig = string(cfgBytes)
		existing.Config = req.Config
	}

	_, err = h.db.Exec(`
		UPDATE targets
		SET name = ?, type = ?, host = ?, port = ?, auth_config = ?, polling_interval = ?
		WHERE id = ? AND user_id = ?
	`, existing.Name, string(existing.Type), existing.Host, existing.Port, existing.AuthConfig, existing.PollingInterval, targetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate target"})
		return
	}

	// Restart poller for this target
	if worker.GlobalScheduler != nil {
		worker.GlobalScheduler.RestartTarget(targetID)
	}

	// Mask for client response
	responseTarget := existing
	if responseTarget.Config != nil {
		maskedCfg := *responseTarget.Config
		maskSensitiveConfig(&maskedCfg)
		responseTarget.Config = &maskedCfg
		maskedJSON, _ := json.Marshal(maskedCfg)
		responseTarget.AuthConfig = string(maskedJSON)
	}

	c.JSON(http.StatusOK, responseTarget)
}

func (h *TargetHandler) DeleteTarget(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target tidak valid"})
		return
	}

	// Verify ownership before stopping poller
	var exists bool
	err = h.db.QueryRow(`SELECT 1 FROM targets WHERE id = ? AND user_id = ?`, targetID, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target tidak ditemukan"})
		return
	}

	// Stop poller
	if worker.GlobalScheduler != nil {
		worker.GlobalScheduler.StopTarget(targetID)
	}

	purgeHistory := c.Query("purge_history") == "true"
	if purgeHistory {
		_, _ = h.db.Exec(`DELETE FROM metrics_raw WHERE target_id = ?`, targetID)
		_, _ = h.db.Exec(`DELETE FROM metrics_hourly WHERE target_id = ?`, targetID)
		_, _ = h.db.Exec(`DELETE FROM metrics_daily WHERE target_id = ?`, targetID)
	}

	_, err = h.db.Exec(`DELETE FROM targets WHERE id = ? AND user_id = ?`, targetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus target"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target berhasil dihapus"})
}

func (h *TargetHandler) TestConnection(c *gin.Context) {
	var req models.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data test koneksi tidak valid"})
		return
	}

	cleanHost := strings.TrimSpace(req.Host)
	if err := validateTarget("Test", req.Type, cleanHost, req.Port, 60); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Host = cleanHost

	resp := collector.TestTargetConnection(&req)
	c.JSON(http.StatusOK, resp)
}
