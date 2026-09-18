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
	"sim_monit/server/models"
)

type MetricHandler struct {
	db *sql.DB
}

func NewMetricHandler(db *sql.DB) *MetricHandler {
	return &MetricHandler{db: db}
}

func (h *MetricHandler) GetTargetMetrics(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target tidak valid"})
		return
	}

	// Verify target ownership to prevent IDOR
	var exists bool
	err = h.db.QueryRow(`SELECT 1 FROM targets WHERE id = ? AND user_id = ?`, targetID, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target tidak ditemukan"})
		return
	}

	rawRange := strings.ToLower(c.DefaultQuery("range", "1h"))
	validRanges := map[string]bool{
		"1h": true, "24h": true, "7d": true, "mtd": true, "ytd": true, "custom": true,
	}
	timeRange := "1h"
	if validRanges[rawRange] {
		timeRange = rawRange
	}
	now := time.Now().UTC()

	var startTime time.Time
	var scale string

	switch timeRange {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
		scale = "Raw Data"
	case "24h":
		startTime = now.Add(-24 * time.Hour)
		scale = "Raw Data"
	case "7d":
		startTime = now.AddDate(0, 0, -7)
		scale = "Raw Data"
	case "mtd":
		// Month to date
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		scale = "Hourly Avg"
	case "ytd":
		// Year to date
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		scale = "Daily Avg"
	case "custom":
		fromStr := c.Query("from")
		toStr := c.Query("to")
		if fromStr != "" {
			t, err := time.Parse(time.RFC3339, fromStr)
			if err == nil {
				startTime = t
			} else {
				t2, err2 := time.Parse("2006-01-02", fromStr)
				if err2 == nil {
					startTime = t2
				} else {
					startTime = now.Add(-24 * time.Hour)
				}
			}
		} else {
			startTime = now.Add(-24 * time.Hour)
		}

		// Check duration for scale decision
		duration := now.Sub(startTime)
		if duration <= 7*24*time.Hour {
			scale = "Raw Data"
		} else if duration <= 90*24*time.Hour {
			scale = "Hourly Avg"
		} else {
			scale = "Daily Avg"
		}
		_ = toStr
	default:
		startTime = now.Add(-1 * time.Hour)
		scale = "Raw Data"
	}

	dataPoints := make([]models.ChartDataPoint, 0)
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	if scale == "Raw Data" {
		rows, err := h.db.Query(`
			SELECT timestamp, latency_ms, cpu_pct, ram_pct, disk_pct, network_speed, raw_details
			FROM metrics_raw
			WHERE target_id = ? AND timestamp >= ?
			ORDER BY timestamp ASC
		`, targetID, startTimeStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var dp models.ChartDataPoint
				var detailsStr string
				var ts string
				if err := rows.Scan(&ts, &dp.LatencyMs, &dp.CPUPct, &dp.RAMPct, &dp.DiskPct, &dp.NetworkSpeed, &detailsStr); err == nil {
					dp.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
					if dp.Timestamp.IsZero() {
						dp.Timestamp, _ = time.Parse(time.RFC3339, ts)
					}
					if detailsStr != "" {
						var detailsMap map[string]any
						if err := json.Unmarshal([]byte(detailsStr), &detailsMap); err == nil {
							dp.Details = detailsMap
						}
					}
					dataPoints = append(dataPoints, dp)
				}
			}
		}
	} else if scale == "Hourly Avg" {
		rows, err := h.db.Query(`
			SELECT timestamp, avg_latency, max_latency, avg_cpu, avg_ram, avg_disk
			FROM metrics_hourly
			WHERE target_id = ? AND timestamp >= ?
			ORDER BY timestamp ASC
		`, targetID, startTimeStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var dp models.ChartDataPoint
				var ts string
				if err := rows.Scan(&ts, &dp.LatencyMs, &dp.MaxLatencyMs, &dp.CPUPct, &dp.RAMPct, &dp.DiskPct); err == nil {
					dp.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
					dataPoints = append(dataPoints, dp)
				}
			}
		}
	} else {
		// Daily Avg
		rows, err := h.db.Query(`
			SELECT timestamp, avg_latency, max_latency, avg_cpu, avg_ram, avg_disk
			FROM metrics_daily
			WHERE target_id = ? AND timestamp >= ?
			ORDER BY timestamp ASC
		`, targetID, startTimeStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var dp models.ChartDataPoint
				var ts string
				if err := rows.Scan(&ts, &dp.LatencyMs, &dp.MaxLatencyMs, &dp.CPUPct, &dp.RAMPct, &dp.DiskPct); err == nil {
					dp.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
					dataPoints = append(dataPoints, dp)
				}
			}
		}
	}

	c.JSON(http.StatusOK, models.MetricsQueryResponse{
		TargetID:  targetID,
		TimeRange: timeRange,
		Scale:     scale,
		Data:      dataPoints,
	})
}

func (h *MetricHandler) GetTargetStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target tidak valid"})
		return
	}

	// 1. Get target basic info (verified with user_id)
	var target models.Target
	var authConfigStr string
	err = h.db.QueryRow(`
		SELECT id, user_id, name, type, host, port, auth_config, polling_interval, status, created_at
		FROM targets
		WHERE id = ? AND user_id = ?
	`, targetID, userID).Scan(&target.ID, &target.UserID, &target.Name, &target.Type, &target.Host, &target.Port, &authConfigStr, &target.PollingInterval, &target.Status, &target.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target tidak ditemukan"})
		return
	}

	// 2. Get latest raw metric
	var latestMetric models.MetricRaw
	var rawDetailsStr string
	var tsStr string
	err = h.db.QueryRow(`
		SELECT id, timestamp, latency_ms, cpu_pct, ram_pct, disk_pct, network_speed, raw_details
		FROM metrics_raw
		WHERE target_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, targetID).Scan(&latestMetric.ID, &tsStr, &latestMetric.LatencyMs, &latestMetric.CPUPct, &latestMetric.RAMPct, &latestMetric.DiskPct, &latestMetric.NetworkSpeed, &rawDetailsStr)

	latestMetric.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
	if latestMetric.Timestamp.IsZero() {
		latestMetric.Timestamp = time.Now()
	}

	var detailsMap map[string]any
	if rawDetailsStr != "" {
		_ = json.Unmarshal([]byte(rawDetailsStr), &detailsMap)
	}

	// 3. Sparkline (Last 20 latency points)
	sparkline := make([]float64, 0)
	sparkRows, err := h.db.Query(`
		SELECT latency_ms
		FROM metrics_raw
		WHERE target_id = ?
		ORDER BY id DESC
		LIMIT 20
	`, targetID)
	if err == nil {
		defer sparkRows.Close()
		temp := make([]float64, 0)
		for sparkRows.Next() {
			var lat float64
			if err := sparkRows.Scan(&lat); err == nil {
				temp = append(temp, lat)
			}
		}
		// Reverse to chronological
		for i := len(temp) - 1; i >= 0; i-- {
			sparkline = append(sparkline, temp[i])
		}
	}

	// 4. Calculate Uptime Rate (%) based on selected time-range per file 02
	rawRange := strings.ToLower(c.DefaultQuery("range", "24h"))
	validRanges := map[string]bool{
		"1h": true, "24h": true, "7d": true, "mtd": true, "ytd": true, "custom": true,
	}
	timeRange := "24h"
	if validRanges[rawRange] {
		timeRange = rawRange
	}
	now := time.Now().UTC()
	var startTime time.Time

	switch timeRange {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "mtd":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "ytd":
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case "custom":
		fromStr := c.Query("from")
		if fromStr != "" {
			if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
				startTime = t
			} else if t2, err2 := time.Parse("2006-01-02", fromStr); err2 == nil {
				startTime = t2
			} else {
				startTime = now.Add(-24 * time.Hour)
			}
		} else {
			startTime = now.Add(-24 * time.Hour)
		}
	default:
		startTime = now.Add(-24 * time.Hour)
	}

	startTimeStr := startTime.Format("2006-01-02 15:04:05")
	var totalChecks, onlineChecks int
	_ = h.db.QueryRow(`
		SELECT 
			COUNT(*), 
			COUNT(CASE WHEN json_extract(raw_details, '$.status') = 'ONLINE' OR (json_extract(raw_details, '$.status') IS NULL AND raw_details NOT LIKE '%"error"%' AND raw_details NOT LIKE '%"port_status":"CLOSED"%') THEN 1 END)
		FROM metrics_raw
		WHERE target_id = ? AND timestamp >= ?
	`, targetID, startTimeStr).Scan(&totalChecks, &onlineChecks)

	uptimeRate := 100.0
	if totalChecks > 0 {
		uptimeRate = (float64(onlineChecks) / float64(totalChecks)) * 100.0
	} else if timeRange == "ytd" || timeRange == "mtd" {
		var avgUptime sql.NullFloat64
		_ = h.db.QueryRow(`
			SELECT AVG(uptime_pct)
			FROM metrics_daily
			WHERE target_id = ? AND timestamp >= ?
		`, targetID, startTimeStr).Scan(&avgUptime)
		if avgUptime.Valid {
			uptimeRate = avgUptime.Float64
		}
	}

	// 5. Calculate real-time duration in current status (e.g. "Up for 14d 6h 23m" or "Down for 5m")
	uptimeDuration := "Active"
	oppositeStatus := "OFFLINE"
	if target.Status != models.TargetStatusOnline {
		oppositeStatus = "ONLINE"
	}

	var lastStateChangeStr sql.NullString
	_ = h.db.QueryRow(`
		SELECT timestamp 
		FROM metrics_raw 
		WHERE target_id = ? AND (json_extract(raw_details, '$.status') = ? OR raw_details LIKE '%"error"%')
		ORDER BY id DESC LIMIT 1
	`, targetID, oppositeStatus).Scan(&lastStateChangeStr)

	var stateDuration time.Duration
	if lastStateChangeStr.Valid {
		tChange, _ := time.Parse("2006-01-02 15:04:05", lastStateChangeStr.String)
		if !tChange.IsZero() {
			stateDuration = time.Since(tChange)
		}
	}
	if stateDuration <= 0 {
		stateDuration = time.Since(target.CreatedAt)
	}

	days := int(stateDuration.Hours() / 24)
	hours := int(stateDuration.Hours()) % 24
	mins := int(stateDuration.Minutes()) % 60
	prefix := "Up for "
	if target.Status != models.TargetStatusOnline {
		prefix = "Down for "
	}

	if days > 0 {
		uptimeDuration = fmt.Sprintf("%s%dd %dh %dm", prefix, days, hours, mins)
	} else if hours > 0 {
		uptimeDuration = fmt.Sprintf("%s%dh %dm", prefix, hours, mins)
	} else if mins > 0 {
		uptimeDuration = fmt.Sprintf("%s%dm", prefix, mins)
	} else {
		uptimeDuration = fmt.Sprintf("%s< 1m", prefix)
	}

	// 6. Build Port Matrix
	portMatrix := make([]models.PortMatrixItem, 0)
	if detailsMap != nil {
		if pm, ok := detailsMap["port_matrix"].([]any); ok {
			for _, item := range pm {
				if itemMap, ok := item.(map[string]any); ok {
					name, _ := itemMap["name"].(string)
					portVal, _ := itemMap["port"].(float64)
					status, _ := itemMap["status"].(string)
					latVal, _ := itemMap["latency_ms"].(float64)
					portMatrix = append(portMatrix, models.PortMatrixItem{
						Name:    name,
						Port:    int(portVal),
						Status:  status,
						Latency: latVal,
					})
				}
			}
		}
	}

	// If no port matrix in details, provide standard items according to target type
	if len(portMatrix) == 0 {
		if target.Type == models.TargetTypeWebsite || target.Type == models.TargetTypeAPI {
			portMatrix = append(portMatrix,
				models.PortMatrixItem{Name: "HTTP/HTTPS", Port: 443, Status: string(target.Status), Latency: latestMetric.LatencyMs},
				models.PortMatrixItem{Name: "DNS Resolution", Port: 53, Status: "RESOLVED", Latency: 4.2},
			)
		} else if target.Type == models.TargetTypeDatabase {
			portMatrix = append(portMatrix,
				models.PortMatrixItem{Name: "DB Port Listener", Port: target.Port, Status: string(target.Status), Latency: latestMetric.LatencyMs},
			)
		}
	}

	var httpStatus int
	if detailsMap != nil {
		if sc, ok := detailsMap["status_code"].(float64); ok {
			httpStatus = int(sc)
		}
	}

	c.JSON(http.StatusOK, models.TargetStatsResponse{
		TargetID:       targetID,
		TimeRange:      timeRange,
		Status:         target.Status,
		UptimeDuration: uptimeDuration,
		CurrentLatency: latestMetric.LatencyMs,
		UptimeRate:     uptimeRate,
		TotalChecks:    totalChecks,
		OnlineChecks:   onlineChecks,
		CPUPct:         latestMetric.CPUPct,
		RAMPct:         latestMetric.RAMPct,
		DiskPct:        latestMetric.DiskPct,
		Sparkline:      sparkline,
		PortMatrix:     portMatrix,
		HTTPStatus:     httpStatus,
		LastUpdated:    latestMetric.Timestamp,
	})
}
