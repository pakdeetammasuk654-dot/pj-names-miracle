package domain

import (
	"context"
	"time"
)

// HealthStatus represents health status detail
type HealthStatus struct {
	ServiceName string            `json:"service_name"`
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// HealthPayload is passed through the Pipeline filters
type HealthPayload struct {
	ServiceName string
	ClientIP    string
	Result      HealthStatus
}

// HealthUseCase interface defining business logic contract
type HealthUseCase interface {
	CheckHealth(ctx context.Context, clientIP string) (*HealthStatus, error)
}
