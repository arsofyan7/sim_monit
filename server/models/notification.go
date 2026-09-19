package models

type TargetAlertConfig struct {
	Enabled bool   `json:"enabled"`
	ChatID  string `json:"chat_id"`
}

type TelegramSettings struct {
	TelegramEnabled  bool                         `json:"telegram_enabled"`
	TelegramBotToken string                       `json:"telegram_bot_token"`
	TelegramMode     string                       `json:"telegram_mode"` // "all" or "per_target"
	TelegramChatID   string                       `json:"telegram_chat_id"`
	TargetMappings   map[string]TargetAlertConfig `json:"target_mappings"`
}

type UpdateTelegramSettingsRequest struct {
	TelegramEnabled  bool                         `json:"telegram_enabled"`
	TelegramBotToken string                       `json:"telegram_bot_token"`
	TelegramMode     string                       `json:"telegram_mode"`
	TelegramChatID   string                       `json:"telegram_chat_id"`
	TargetMappings   map[string]TargetAlertConfig `json:"target_mappings"`
}

type TestTelegramRequest struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}
