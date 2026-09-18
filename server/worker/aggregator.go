package worker

import (
	"database/sql"
	"log"
	"time"
)

type Aggregator struct {
	db *sql.DB
}

func InitAggregator(db *sql.DB) *Aggregator {
	return &Aggregator{db: db}
}

func (a *Aggregator) Start() {
	log.Println("[Aggregator] Starting Hourly & Daily background roll-up workers...")

	// Hourly aggregation worker
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.AggregateHourly()
			}
		}
	}()

	// Daily aggregation worker
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.AggregateDaily()
			}
		}
	}()

	// Run an initial quick pass
	go func() {
		time.Sleep(5 * time.Second)
		a.AggregateHourly()
		a.AggregateDaily()
	}()
}

// AggregateHourly aggregates raw metrics from the previous hour into metrics_hourly
func (a *Aggregator) AggregateHourly() {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour).Format("2006-01-02 15:00:00")
	currentHour := now.Format("2006-01-02 15:00:00")

	query := `
		INSERT OR IGNORE INTO metrics_hourly (target_id, timestamp, avg_latency, max_latency, avg_cpu, avg_ram, avg_disk)
		SELECT 
			target_id,
			strftime('%Y-%m-%d %H:00:00', timestamp) as hour_slot,
			AVG(latency_ms) as avg_latency,
			MAX(latency_ms) as max_latency,
			AVG(cpu_pct) as avg_cpu,
			AVG(ram_pct) as avg_ram,
			AVG(disk_pct) as avg_disk
		FROM metrics_raw
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY target_id, hour_slot
	`

	res, err := a.db.Exec(query, oneHourAgo, currentHour)
	if err != nil {
		log.Printf("[Aggregator] Error running hourly aggregation: %v\n", err)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Aggregator] Hourly aggregation processed %d records\n", rowsAffected)
	}
}

// AggregateDaily aggregates hourly metrics into metrics_daily with uptime percentage
func (a *Aggregator) AggregateDaily() {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour).Format("2006-01-02 00:00:00")
	today := now.Format("2006-01-02 00:00:00")

	// Query metrics_raw directly for the 24-hour day to compute exact SLA uptime percentage
	query := `
		INSERT OR IGNORE INTO metrics_daily (target_id, timestamp, avg_latency, max_latency, avg_cpu, avg_ram, avg_disk, uptime_pct)
		SELECT 
			target_id,
			strftime('%Y-%m-%d 00:00:00', timestamp) as day_slot,
			AVG(latency_ms) as avg_latency,
			MAX(latency_ms) as max_latency,
			AVG(cpu_pct) as avg_cpu,
			AVG(ram_pct) as avg_ram,
			AVG(disk_pct) as avg_disk,
			ROUND(COALESCE(100.0 * COUNT(CASE WHEN json_extract(raw_details, '$.status') = 'ONLINE' OR (json_extract(raw_details, '$.status') IS NULL AND raw_details NOT LIKE '%"error"%') THEN 1 END) / NULLIF(COUNT(*), 0), 100.0), 2) as uptime_pct
		FROM metrics_raw
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY target_id, day_slot
	`

	res, err := a.db.Exec(query, yesterday, today)
	if err != nil {
		log.Printf("[Aggregator] Error running daily aggregation: %v\n", err)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Aggregator] Daily aggregation processed %d records\n", rowsAffected)
	}
}
