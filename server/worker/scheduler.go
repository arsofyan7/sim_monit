package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"sim_monit/server/alert"
	"sim_monit/server/collector"
	"sim_monit/server/models"
)

type Scheduler struct {
	db         *sql.DB
	mu         sync.Mutex
	stopChan   map[int64]chan struct{}
	lastStatus map[int64]models.TargetStatus
	downSince  map[int64]time.Time
}

var GlobalScheduler *Scheduler

func InitScheduler(db *sql.DB) *Scheduler {
	s := &Scheduler{
		db:         db,
		stopChan:   make(map[int64]chan struct{}),
		lastStatus: make(map[int64]models.TargetStatus),
		downSince:  make(map[int64]time.Time),
	}
	GlobalScheduler = s
	return s
}

func (s *Scheduler) Start() {
	log.Println("[Scheduler] Starting agentless polling engine...")
	s.ReloadTargets()

	// Periodic checker to ensure all DB targets have active pollers
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			s.ReloadTargets()
		}
	}()
}

func (s *Scheduler) ReloadTargets() {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT id, user_id, name, type, host, port, auth_config, polling_interval, status, created_at FROM targets`)
	if err != nil {
		log.Printf("[Scheduler] Error fetching targets: %v\n", err)
		return
	}
	defer rows.Close()

	activeTargetIDs := make(map[int64]bool)

	for rows.Next() {
		var t models.Target
		var authConfigStr string
		err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Type, &t.Host, &t.Port, &authConfigStr, &t.PollingInterval, &t.Status, &t.CreatedAt)
		if err != nil {
			continue
		}
		t.AuthConfig = authConfigStr
		var cfg models.AuthConfig
		if err := json.Unmarshal([]byte(authConfigStr), &cfg); err == nil {
			t.Config = &cfg
		}

		activeTargetIDs[t.ID] = true

		// If this target isn't already running a poller goroutine, start it
		if _, exists := s.stopChan[t.ID]; !exists {
			stopCh := make(chan struct{})
			s.stopChan[t.ID] = stopCh
			go s.pollTarget(t, stopCh)
		}
	}

	// Stop pollers for targets that were deleted
	for id, stopCh := range s.stopChan {
		if !activeTargetIDs[id] {
			close(stopCh)
			delete(s.stopChan, id)
			log.Printf("[Scheduler] Stopped poller for removed target ID %d\n", id)
		}
	}
}

func (s *Scheduler) StopTarget(targetID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stopCh, exists := s.stopChan[targetID]; exists {
		close(stopCh)
		delete(s.stopChan, targetID)
		log.Printf("[Scheduler] Poller stopped for target ID %d\n", targetID)
	}
}

func (s *Scheduler) RestartTarget(targetID int64) {
	s.StopTarget(targetID)
	s.ReloadTargets()
}

func (s *Scheduler) pollTarget(target models.Target, stopCh chan struct{}) {
	intervalSec := target.PollingInterval
	if intervalSec < 5 {
		intervalSec = 10
	}
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	log.Printf("[Scheduler] Started poller for target '%s' (ID: %d, Type: %s, Interval: %ds)\n",
		target.Name, target.ID, target.Type, intervalSec)

	// Execute initial check immediately
	s.runSingleCollection(&target)

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			s.runSingleCollection(&target)
		}
	}
}

func (s *Scheduler) runSingleCollection(target *models.Target) {
	metric, status, collErr := collector.CollectTarget(target)
	if metric == nil {
		metric = &models.MetricRaw{
			TargetID:  target.ID,
			Timestamp: time.Now(),
		}
	}

	// Embed explicit status flag in raw_details for 100% reliable SLA calculations
	var detailsMap map[string]any
	if metric.RawDetails != "" {
		_ = json.Unmarshal([]byte(metric.RawDetails), &detailsMap)
	}
	if detailsMap == nil {
		detailsMap = make(map[string]any)
	}
	detailsMap["status"] = string(status)
	detailsMap["is_up"] = (status == models.TargetStatusOnline)
	if updatedJSON, err := json.Marshal(detailsMap); err == nil {
		metric.RawDetails = string(updatedJSON)
	}

	// Save raw metric to SQLite
	_, err := s.db.Exec(`
		INSERT INTO metrics_raw (target_id, timestamp, latency_ms, cpu_pct, ram_pct, disk_pct, network_speed, raw_details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, metric.TargetID, time.Now().Format("2006-01-02 15:04:05"), metric.LatencyMs, metric.CPUPct, metric.RAMPct, metric.DiskPct, metric.NetworkSpeed, metric.RawDetails)

	if err != nil {
		log.Printf("[Scheduler] Failed to insert raw metric for target %d: %v\n", target.ID, err)
	}

	// Detect status transitions for Telegram alerting & Incident tracking
	s.mu.Lock()
	prevStatus, hasPrev := s.lastStatus[target.ID]
	if !hasPrev {
		prevStatus = target.Status
	}
	s.lastStatus[target.ID] = status
	downTime, hasDownTime := s.downSince[target.ID]
	s.mu.Unlock()

	// 1. Transisi ONLINE/PENDING -> OFFLINE (DOWN Alert & Buat Record Insiden)
	if status == models.TargetStatusOffline && prevStatus != models.TargetStatusOffline {
		now := time.Now()
		s.mu.Lock()
		s.downSince[target.ID] = now
		s.mu.Unlock()

		causeMsg := "Koneksi terputus / unreachable"
		if collErr != nil {
			causeMsg = collErr.Error()
		}

		// Insert record insiden ke SQLite
		_, err := s.db.Exec(`
			INSERT INTO incidents (target_id, started_at, resolved_at, duration_seconds, cause)
			VALUES (?, ?, NULL, 0, ?)
		`, target.ID, now.Format("2006-01-02 15:04:05"), causeMsg)
		if err != nil {
			log.Printf("[Incident] Gagal mencatat insiden baru untuk target %d: %v\n", target.ID, err)
		}

		go s.dispatchAlert(target, models.TargetStatusOffline, metric, collErr, 0)
	} else if status == models.TargetStatusOnline && prevStatus == models.TargetStatusOffline {
		// 2. Transisi OFFLINE -> ONLINE (RECOVERY Alert & Selesaikan Record Insiden)
		now := time.Now()
		var downtime time.Duration
		if hasDownTime {
			downtime = time.Since(downTime)
		}
		s.mu.Lock()
		delete(s.downSince, target.ID)
		s.mu.Unlock()

		nowStr := now.Format("2006-01-02 15:04:05")
		// Update record insiden yang masih terbuka (resolved_at IS NULL)
		_, err := s.db.Exec(`
			UPDATE incidents
			SET resolved_at = ?,
			    duration_seconds = MAX(1, CAST((strftime('%s', ?) - strftime('%s', started_at)) AS INTEGER))
			WHERE target_id = ? AND resolved_at IS NULL
		`, nowStr, nowStr, target.ID)
		if err != nil {
			log.Printf("[Incident] Gagal memperbarui status selesai insiden target %d: %v\n", target.ID, err)
		}

		go s.dispatchAlert(target, models.TargetStatusOnline, metric, nil, downtime)
	}

	// Update status in targets table if changed
	_, _ = s.db.Exec(`UPDATE targets SET status = ? WHERE id = ?`, string(status), target.ID)
}

func (s *Scheduler) dispatchAlert(target *models.Target, newStatus models.TargetStatus, metric *models.MetricRaw, collErr error, downtime time.Duration) {
	var (
		enabled      int
		botToken     string
		mode         string
		globalChatID string
		mappingsJSON string
	)

	err := s.db.QueryRow(`
		SELECT telegram_enabled, telegram_bot_token, telegram_mode, telegram_chat_id, target_mappings
		FROM notification_settings
		WHERE user_id = ?
	`, target.UserID).Scan(&enabled, &botToken, &mode, &globalChatID, &mappingsJSON)

	if err != nil || enabled != 1 {
		return // Notifikasi dinonaktifkan
	}

	botToken = strings.TrimSpace(botToken)
	if botToken == "" {
		// Sesuai requirement: jika toggle aktif tapi bot token kosong, abaikan agar tidak error
		return
	}

	var chatIDToSend string

	if mode == "per_target" {
		mappings := make(map[string]models.TargetAlertConfig)
		if mappingsJSON != "" {
			_ = json.Unmarshal([]byte(mappingsJSON), &mappings)
		}
		targetKey := fmt.Sprintf("%d", target.ID)
		if cfg, exists := mappings[targetKey]; exists && cfg.Enabled {
			chatIDToSend = strings.TrimSpace(cfg.ChatID)
		} else {
			// Mode per-target aktif tapi target ini tidak dicentang/diaktifkan
			return
		}
	} else {
		// Mode global ("all")
		chatIDToSend = strings.TrimSpace(globalChatID)
	}

	if chatIDToSend == "" {
		// Sesuai requirement: jika chat ID kosong, abaikan agar tidak error
		return
	}

	var msgText string
	now := time.Now()

	if newStatus == models.TargetStatusOffline {
		errMsg := ""
		if collErr != nil {
			errMsg = collErr.Error()
		}
		msgText = alert.FormatDownAlert(target, errMsg, now)
	} else if newStatus == models.TargetStatusOnline {
		lat := 0.0
		if metric != nil {
			lat = metric.LatencyMs
		}
		msgText = alert.FormatRecoveryAlert(target, lat, downtime, now)
	} else {
		return
	}

	if err := alert.SendTelegramMessage(botToken, chatIDToSend, msgText); err != nil {
		log.Printf("[Alert] Gagal mengirim Telegram alert untuk target %s (ID %d): %v\n", target.Name, target.ID, err)
	} else {
		log.Printf("[Alert] Berhasil mengirim Telegram alert untuk target %s (ID %d, Status: %s) ke Chat ID: %s\n", target.Name, target.ID, newStatus, chatIDToSend)
	}
}

func (s *Scheduler) TriggerImmediatePoll(target *models.Target) {
	go s.runSingleCollection(target)
}
