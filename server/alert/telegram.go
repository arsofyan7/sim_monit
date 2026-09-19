package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sim_monit/server/models"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SendTelegramMessage sends a message to Telegram Bot API.
// If botToken or chatID is empty, it silently ignores to prevent errors.
func SendTelegramMessage(botToken, chatID, text string) error {
	token := strings.TrimSpace(botToken)
	chat := strings.TrimSpace(chatID)

	if token == "" || chat == "" {
		// Abaikan jika token atau chat id kosong agar tidak menimbulkan error
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chat,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dj %dm %dd", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %dd", m, s)
	}
	return fmt.Sprintf("%d detik", s)
}

func FormatDownAlert(target *models.Target, errReason string, eventTime time.Time) string {
	endpoint := target.Host
	if target.Port > 0 {
		endpoint = fmt.Sprintf("%s:%d", target.Host, target.Port)
	}

	reason := "Tidak dapat dihubungi (Connection failed/timeout)"
	if strings.TrimSpace(errReason) != "" {
		reason = errReason
	}

	return fmt.Sprintf(
		"🔴 <b>[SIM_MONIT ALERT] TARGET DOWN!</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"<b>Target:</b> %s (<i>%s</i>)\n"+
			"<b>Endpoint:</b> <code>%s</code>\n"+
			"<b>Status:</b> <b>OFFLINE</b>\n"+
			"<b>Penyebab:</b> %s\n"+
			"<b>Waktu:</b> %s\n",
		target.Name,
		strings.ToUpper(string(target.Type)),
		endpoint,
		reason,
		eventTime.Format("02 Jan 2006, 15:04:05 MST"),
	)
}

func FormatRecoveryAlert(target *models.Target, latencyMs float64, downtime time.Duration, eventTime time.Time) string {
	endpoint := target.Host
	if target.Port > 0 {
		endpoint = fmt.Sprintf("%s:%d", target.Host, target.Port)
	}

	latencyStr := fmt.Sprintf("%.1f ms", latencyMs)
	if latencyMs < 1 && latencyMs > 0 {
		latencyStr = fmt.Sprintf("%.2f ms", latencyMs)
	}

	downtimeStr := FormatDuration(downtime)
	if downtime <= 0 {
		downtimeStr = "< 1 menit"
	}

	return fmt.Sprintf(
		"🟢 <b>[SIM_MONIT RECOVERY] TARGET BACK ONLINE!</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"<b>Target:</b> %s (<i>%s</i>)\n"+
			"<b>Endpoint:</b> <code>%s</code>\n"+
			"<b>Status:</b> <b>ONLINE</b> (Latency: %s)\n"+
			"<b>Total Downtime:</b> %s\n"+
			"<b>Waktu Pulih:</b> %s\n",
		target.Name,
		strings.ToUpper(string(target.Type)),
		endpoint,
		latencyStr,
		downtimeStr,
		eventTime.Format("02 Jan 2006, 15:04:05 MST"),
	)
}

func FormatTestMessage() string {
	return "🧪 <b>[SIM_MONIT] Uji Coba Notifikasi Telegram Berhasil!</b>\n" +
		"━━━━━━━━━━━━━━━━━━━━\n" +
		"Koneksi antara bot Telegram dan SIM_MONIT berhasil diverifikasi.\n" +
		"Pesan alert status target akan dikirimkan ke chat ini secara otomatis saat terjadi gangguan atau pemulihan."
}
