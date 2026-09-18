package collector

import (
	"encoding/json"
	"fmt"

	"sim_monit/server/models"
)

func CollectTarget(target *models.Target) (*models.MetricRaw, models.TargetStatus, error) {
	if target.Config == nil && target.AuthConfig != "" {
		var cfg models.AuthConfig
		if err := json.Unmarshal([]byte(target.AuthConfig), &cfg); err == nil {
			target.Config = &cfg
		}
	}

	switch target.Type {
	case models.TargetTypeServer:
		return CollectSSH(target)
	case models.TargetTypeWebsite, models.TargetTypeAPI:
		return CollectHTTP(target)
	case models.TargetTypeDatabase:
		return CollectTCP(target)
	default:
		// Fallback to TCP ping if port > 0, otherwise HTTP
		if target.Port > 0 {
			return CollectTCP(target)
		}
		return CollectHTTP(target)
	}
}

func TestTargetConnection(req *models.TestConnectionRequest) *models.TestConnectionResponse {
	target := &models.Target{
		Type:   req.Type,
		Host:   req.Host,
		Port:   req.Port,
		Config: &req.Config,
	}

	metric, status, err := CollectTarget(target)
	if err != nil && status == models.TargetStatusOffline {
		latency := 0.0
		if metric != nil {
			latency = metric.LatencyMs
		}
		return &models.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
			Latency: latency,
		}
	}

	var detailsMap map[string]any
	if metric != nil && metric.RawDetails != "" {
		_ = json.Unmarshal([]byte(metric.RawDetails), &detailsMap)
	}

	return &models.TestConnectionResponse{
		Success: true,
		Message: "Connection successful!",
		Latency: metric.LatencyMs,
		Details: detailsMap,
	}
}
