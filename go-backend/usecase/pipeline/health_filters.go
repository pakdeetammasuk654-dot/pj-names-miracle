package pipeline

import (
	"fmt"
	"log"
	"time"

	"go-backend/domain"
)

var startTime = time.Now()

// InitHealthFilter initializes basic payload details
type InitHealthFilter struct{}

func (f *InitHealthFilter) Execute(pCtx *domain.PipelineContext[*domain.HealthPayload]) error {
	pCtx.Data.Result = domain.HealthStatus{
		ServiceName: pCtx.Data.ServiceName,
		Status:      "UP",
		Timestamp:   time.Now(),
		Metadata:    make(map[string]string),
	}
	return nil
}

// LoggingFilter handles step-by-step logging filter
type LoggingFilter struct{}

func (f *LoggingFilter) Execute(pCtx *domain.PipelineContext[*domain.HealthPayload]) error {
	log.Printf("[Pipeline Filter] Processing health check request from IP: %s", pCtx.Data.ClientIP)
	return nil
}

// SystemCheckFilter verifies system health metrics and calculates uptime
type SystemCheckFilter struct{}

func (f *SystemCheckFilter) Execute(pCtx *domain.PipelineContext[*domain.HealthPayload]) error {
	uptime := time.Since(startTime).Round(time.Second).String()
	pCtx.Data.Result.Uptime = uptime
	pCtx.Data.Result.Metadata["environment"] = "production"
	pCtx.Data.Result.Metadata["go_version"] = "go1.25"
	return nil
}

// ValidationFilter validates client input IP or params
type ValidationFilter struct{}

func (f *ValidationFilter) Execute(pCtx *domain.PipelineContext[*domain.HealthPayload]) error {
	if pCtx.Data.ClientIP == "blocked" {
		return fmt.Errorf("client IP %s is blocked by security policy", pCtx.Data.ClientIP)
	}
	return nil
}
