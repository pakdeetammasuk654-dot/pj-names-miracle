package usecase

import (
	"context"

	"go-backend/domain"
	"go-backend/usecase/pipeline"
)

type healthUseCase struct {
	serviceName string
}

func NewHealthUseCase(serviceName string) domain.HealthUseCase {
	return &healthUseCase{
		serviceName: serviceName,
	}
}

func (u *healthUseCase) CheckHealth(ctx context.Context, clientIP string) (*domain.HealthStatus, error) {
	payload := &domain.HealthPayload{
		ServiceName: u.serviceName,
		ClientIP:    clientIP,
	}

	pCtx := &domain.PipelineContext[*domain.HealthPayload]{
		Ctx:  ctx,
		Data: payload,
	}

	// Sequential Pipeline Construction (Pipes & Filters Pattern)
	pipe := domain.NewPipeline[*domain.HealthPayload](
		&pipeline.LoggingFilter{},
		&pipeline.ValidationFilter{},
		&pipeline.InitHealthFilter{},
		&pipeline.SystemCheckFilter{},
	)

	// Execute pipeline steps sequentially
	if err := pipe.Execute(pCtx); err != nil {
		return nil, err
	}

	return &pCtx.Data.Result, nil
}
