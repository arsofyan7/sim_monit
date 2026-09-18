package collector

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"sim_monit/server/models"
)

func CollectTCP(target *models.Target) (*models.MetricRaw, models.TargetStatus, error) {
	host := target.Host
	port := target.Port
	if port <= 0 {
		// Default ports by target type
		if target.Type == models.TargetTypeDatabase {
			port = 5432 // Default postgres / generic DB
		} else {
			port = 80
		}
	}

	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	var detailsJSON []byte
	var status models.TargetStatus

	if err != nil {
		status = models.TargetStatusOffline
		detailsJSON, _ = json.Marshal(map[string]any{
			"error":       err.Error(),
			"port_status": "CLOSED",
			"address":     address,
		})
	} else {
		_ = conn.Close()
		status = models.TargetStatusOnline
		detailsJSON, _ = json.Marshal(map[string]any{
			"port_status": "LISTENING",
			"address":     address,
			"port":        port,
		})
	}

	metric := &models.MetricRaw{
		TargetID:   target.ID,
		Timestamp:  time.Now(),
		LatencyMs:  latency,
		RawDetails: string(detailsJSON),
	}

	return metric, status, err
}
