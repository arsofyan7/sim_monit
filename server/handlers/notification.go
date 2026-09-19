package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sim_monit/server/alert"
	"sim_monit/server/models"
)

type NotificationHandler struct {
	db *sql.DB
}

func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

func (h *NotificationHandler) GetTelegramSettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var (
		enabled        int
		botToken       string
		mode           string
		chatID         string
		mappingsJSON   string
		updatedAt      time.Time
	)

	err := h.db.QueryRow(`
		SELECT telegram_enabled, telegram_bot_token, telegram_mode, telegram_chat_id, target_mappings, updated_at
		FROM notification_settings
		WHERE user_id = ?
	`, userID).Scan(&enabled, &botToken, &mode, &chatID, &mappingsJSON, &updatedAt)

	if err == sql.ErrNoRows {
		// Default config: non-aktif
		c.JSON(http.StatusOK, models.TelegramSettings{
			TelegramEnabled:  false,
			TelegramBotToken: "",
			TelegramMode:     "all",
			TelegramChatID:   "",
			TargetMappings:   make(map[string]models.TargetAlertConfig),
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pengaturan notifikasi: " + err.Error()})
		return
	}

	mappings := make(map[string]models.TargetAlertConfig)
	if mappingsJSON != "" {
		_ = json.Unmarshal([]byte(mappingsJSON), &mappings)
	}

	c.JSON(http.StatusOK, models.TelegramSettings{
		TelegramEnabled:  enabled == 1,
		TelegramBotToken: botToken,
		TelegramMode:     mode,
		TelegramChatID:   chatID,
		TargetMappings:   mappings,
	})
}

func (h *NotificationHandler) UpdateTelegramSettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.UpdateTelegramSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid: " + err.Error()})
		return
	}

	enabledInt := 0
	if req.TelegramEnabled {
		enabledInt = 1
	}

	mode := strings.TrimSpace(req.TelegramMode)
	if mode != "per_target" {
		mode = "all"
	}

	if req.TargetMappings == nil {
		req.TargetMappings = make(map[string]models.TargetAlertConfig)
	}

	mappingsBytes, err := json.Marshal(req.TargetMappings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses target mappings: " + err.Error()})
		return
	}

	cleanToken := strings.TrimSpace(req.TelegramBotToken)
	cleanChatID := strings.TrimSpace(req.TelegramChatID)

	_, err = h.db.Exec(`
		INSERT INTO notification_settings (user_id, telegram_enabled, telegram_bot_token, telegram_mode, telegram_chat_id, target_mappings, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			telegram_enabled = excluded.telegram_enabled,
			telegram_bot_token = excluded.telegram_bot_token,
			telegram_mode = excluded.telegram_mode,
			telegram_chat_id = excluded.telegram_chat_id,
			target_mappings = excluded.target_mappings,
			updated_at = excluded.updated_at
	`, userID, enabledInt, cleanToken, mode, cleanChatID, string(mappingsBytes), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengaturan notifikasi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pengaturan notifikasi Telegram berhasil disimpan",
		"data": models.TelegramSettings{
			TelegramEnabled:  req.TelegramEnabled,
			TelegramBotToken: cleanToken,
			TelegramMode:     mode,
			TelegramChatID:   cleanChatID,
			TargetMappings:   req.TargetMappings,
		},
	})
}

func (h *NotificationHandler) TestTelegramNotification(c *gin.Context) {
	var req models.TestTelegramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tes tidak valid"})
		return
	}

	botToken := strings.TrimSpace(req.BotToken)
	chatID := strings.TrimSpace(req.ChatID)

	if botToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bot Token Telegram tidak boleh kosong"})
		return
	}

	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chat ID Telegram tidak boleh kosong"})
		return
	}

	testMsg := alert.FormatTestMessage()
	err := alert.SendTelegramMessage(botToken, chatID, testMsg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal mengirim pesan tes: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesan tes berhasil dikirim ke Telegram!",
	})
}
