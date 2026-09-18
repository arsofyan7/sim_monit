package models

import "time"

type MetricRaw struct {
	ID           int64     `json:"id"`
	TargetID     int64     `json:"target_id"`
	Timestamp    time.Time `json:"timestamp"`
	LatencyMs    float64   `json:"latency_ms"`
	CPUPct       float64   `json:"cpu_pct"`
	RAMPct       float64   `json:"ram_pct"`
	DiskPct      float64   `json:"disk_pct"`
	NetworkSpeed float64   `json:"network_speed"`
	RawDetails   string    `json:"raw_details"` // JSON string
}

type MetricHourly struct {
	ID         int64     `json:"id"`
	TargetID   int64     `json:"target_id"`
	Timestamp  time.Time `json:"timestamp"`
	AvgLatency float64   `json:"avg_latency"`
	MaxLatency float64   `json:"max_latency"`
	AvgCPU     float64   `json:"avg_cpu"`
	AvgRAM     float64   `json:"avg_ram"`
	AvgDisk    float64   `json:"avg_disk"`
}

type MetricDaily struct {
	ID         int64     `json:"id"`
	TargetID   int64     `json:"target_id"`
	Timestamp  time.Time `json:"timestamp"`
	AvgLatency float64   `json:"avg_latency"`
	MaxLatency float64   `json:"max_latency"`
	AvgCPU     float64   `json:"avg_cpu"`
	AvgRAM     float64   `json:"avg_ram"`
	AvgDisk    float64   `json:"avg_disk"`
	UptimePct  float64   `json:"uptime_pct"`
}

// Unified metric point returned to chart
type ChartDataPoint struct {
	Timestamp    time.Time      `json:"timestamp"`
	LatencyMs    float64        `json:"latency_ms"`
	MaxLatencyMs float64        `json:"max_latency_ms,omitempty"`
	CPUPct       float64        `json:"cpu_pct"`
	RAMPct       float64        `json:"ram_pct"`
	DiskPct      float64        `json:"disk_pct"`
	NetworkSpeed float64        `json:"network_speed"`
	Details      map[string]any `json:"details,omitempty"`
}

type MetricsQueryResponse struct {
	TargetID  int64            `json:"target_id"`
	TimeRange string           `json:"time_range"`
	Scale     string           `json:"scale"` // "Raw Data", "Hourly Avg", "Daily Avg"
	Data      []ChartDataPoint `json:"data"`
}

type PortMatrixItem struct {
	Name       string  `json:"name"`
	Port       int     `json:"port"`
	Status     string  `json:"status"` // "CONNECTED", "LISTENING", "CLOSED", "UNREACHABLE"
	Latency    float64 `json:"latency_ms"`
	LatencyAlt float64 `json:"latency"`
}

type TargetStatsResponse struct {
	TargetID       int64            `json:"target_id"`
	TimeRange      string           `json:"time_range"`
	Status         TargetStatus     `json:"status"`
	UptimeDuration string           `json:"uptime_duration"`
	CurrentLatency float64          `json:"current_latency_ms"`
	UptimeRate     float64          `json:"uptime_rate_pct"`
	TotalChecks    int              `json:"total_checks"`
	OnlineChecks   int              `json:"online_checks"`
	CPUPct         float64          `json:"cpu_pct"`
	RAMPct         float64          `json:"ram_pct"`
	DiskPct        float64          `json:"disk_pct"`
	Sparkline      []float64        `json:"sparkline"`
	PortMatrix     []PortMatrixItem `json:"port_matrix"`
	PortMatrixAlt  []PortMatrixItem `json:"portMatrix"`
	HTTPStatus     int              `json:"http_status,omitempty"`
	LastUpdated    time.Time        `json:"last_updated"`
}
