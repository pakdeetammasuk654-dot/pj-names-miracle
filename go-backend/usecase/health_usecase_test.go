package usecase_test

import (
	"context"
	"testing"

	"go-backend/usecase"
)

func TestHealthUseCase_CheckHealth_PipelineSuccess(t *testing.T) {
	uc := usecase.NewHealthUseCase("Test-Service")

	status, err := uc.CheckHealth(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if status.Status != "UP" {
		t.Errorf("expected status UP, got %s", status.Status)
	}

	if status.ServiceName != "Test-Service" {
		t.Errorf("expected ServiceName Test-Service, got %s", status.ServiceName)
	}

	if status.Uptime == "" {
		t.Errorf("expected uptime to be calculated by SystemCheckFilter")
	}
}

func TestHealthUseCase_CheckHealth_PipelineValidationError(t *testing.T) {
	uc := usecase.NewHealthUseCase("Test-Service")

	_, err := uc.CheckHealth(context.Background(), "blocked")
	if err == nil {
		t.Fatalf("expected validation error for blocked IP, got nil")
	}
}
