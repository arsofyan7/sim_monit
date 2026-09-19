package models

import "time"

type Incident struct {
	ID              int64      `json:"id"`
	TargetID        int64      `json:"target_id"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	DurationSeconds int64      `json:"duration_seconds"`
	Cause           string     `json:"cause"`
}
