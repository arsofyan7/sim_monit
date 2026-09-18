package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// PRAGMA Optimizations per file 03
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
	}

	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			return nil, fmt.Errorf("failed to execute pragma '%s': %w", p, err)
		}
	}

	// Connection pool tuning for SQLite WAL
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)

	if err := migrateSchema(conn); err != nil {
		return nil, fmt.Errorf("failed to migrate db schema: %w", err)
	}

	DB = conn
	log.Println("[DB] SQLite database initialized successfully with PRAGMA WAL at:", dbPath)
	return DB, nil
}

func migrateSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL, -- 'server', 'website', 'api', 'database'
		host TEXT NOT NULL,
		port INTEGER DEFAULT 0,
		auth_config TEXT DEFAULT '{}',
		polling_interval INTEGER DEFAULT 60, -- seconds (10, 30, 60, 300)
		status TEXT DEFAULT 'PENDING', -- 'ONLINE', 'OFFLINE', 'PENDING'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS metrics_raw (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		latency_ms REAL DEFAULT 0,
		cpu_pct REAL DEFAULT 0,
		ram_pct REAL DEFAULT 0,
		disk_pct REAL DEFAULT 0,
		network_speed REAL DEFAULT 0,
		raw_details TEXT DEFAULT '{}',
		FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS metrics_hourly (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		avg_latency REAL DEFAULT 0,
		max_latency REAL DEFAULT 0,
		avg_cpu REAL DEFAULT 0,
		avg_ram REAL DEFAULT 0,
		avg_disk REAL DEFAULT 0,
		FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS metrics_daily (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		avg_latency REAL DEFAULT 0,
		max_latency REAL DEFAULT 0,
		avg_cpu REAL DEFAULT 0,
		avg_ram REAL DEFAULT 0,
		avg_disk REAL DEFAULT 0,
		uptime_pct REAL DEFAULT 100.0,
		FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
	);

	-- Indices per spec in file 03
	CREATE INDEX IF NOT EXISTS idx_metrics_raw_target_time ON metrics_raw(target_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_hourly_target_time ON metrics_hourly(target_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_daily_target_time ON metrics_daily(target_id, timestamp);
	`

	_, err := db.Exec(schema)
	return err
}
