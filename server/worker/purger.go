package worker

import (
	"database/sql"
	"log"
	"time"
)

type Purger struct {
	db *sql.DB
}

func InitPurger(db *sql.DB) *Purger {
	return &Purger{db: db}
}

func (p *Purger) Start() {
	log.Println("[Purger] Starting Auto-Purge background retention cleaner...")

	go func() {
		// Run every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		// Run once on startup
		p.PurgeOldMetrics()

		for range ticker.C {
			p.PurgeOldMetrics()
		}
	}()
}

func (p *Purger) PurgeOldMetrics() {
	now := time.Now().UTC()

	// 1. Purge metrics_raw older than 7 days
	sevenDaysAgo := now.AddDate(0, 0, -7).Format("2006-01-02 15:04:05")
	resRaw, err := p.db.Exec(`DELETE FROM metrics_raw WHERE timestamp < ?`, sevenDaysAgo)
	if err != nil {
		log.Printf("[Purger] Error purging metrics_raw: %v\n", err)
	} else if count, _ := resRaw.RowsAffected(); count > 0 {
		log.Printf("[Purger] Purged %d raw metrics older than 7 days\n", count)
	}

	// 2. Purge metrics_hourly older than 90 days
	ninetyDaysAgo := now.AddDate(0, 0, -90).Format("2006-01-02 15:04:05")
	resHourly, err := p.db.Exec(`DELETE FROM metrics_hourly WHERE timestamp < ?`, ninetyDaysAgo)
	if err != nil {
		log.Printf("[Purger] Error purging metrics_hourly: %v\n", err)
	} else if count, _ := resHourly.RowsAffected(); count > 0 {
		log.Printf("[Purger] Purged %d hourly metrics older than 90 days\n", count)
	}
}
