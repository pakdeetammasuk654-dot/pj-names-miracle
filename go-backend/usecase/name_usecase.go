package usecase

import (
	"context"
	"strings"

	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase/pipeline"
)

type nameSearchUseCase struct {
	repo repository.NameRepository
}

func NewNameSearchUseCase(repo repository.NameRepository) domain.NameSearchUseCase {
	return &nameSearchUseCase{repo: repo}
}

func (u *nameSearchUseCase) SearchTopNames(ctx context.Context, req domain.NameSearchRequest) ([]domain.NameResultItem, error) {
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Mode == "" {
		req.Mode = domain.ModeStandalone
	}

	payload := &domain.NameSearchPayload{Request: req}
	if req.Mode == domain.ModeCombined && strings.TrimSpace(req.Surname) != "" {
		satSum, shaSum, err := u.repo.GetSurnameSums(ctx, req.Surname)
		if err != nil {
			return nil, err
		}
		payload.SurnameSatSum = satSum
		payload.SurnameShaSum = shaSum
	}

	pCtx := &domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  ctx,
		Data: payload,
	}

	// Sequential Pipeline Pattern (Extensible Pipes & Filters)
	pipe := domain.NewPipeline[*domain.NameSearchPayload](
		pipeline.NewModeFilter(u.repo), // Step 1: Filter D10/D8/D5 by Search Mode
		pipeline.NewKalakiniFilter(),   // Step 2: Day of Birth Kalakini Safety Filter (100% Kalakini Free)
		pipeline.NewScoringFilter(),    // Step Final: Scoring & Top 100 Ranking
	)

	if err := pipe.Execute(pCtx); err != nil {
		return nil, err
	}

	return pCtx.Data.FinalResults, nil
}

func (u *nameSearchUseCase) GetAllNumbersMap(ctx context.Context) (map[int]domain.NumberMeaning, error) {
	return u.repo.GetAllNumbersMap(ctx)
}

func (u *nameSearchUseCase) GetSumsByTHName(ctx context.Context, thname string) (int, int, string, error) {
	return u.repo.GetSumsByTHName(ctx, thname)
}

func (u *nameSearchUseCase) GetSumsForText(ctx context.Context, text string) (int, int, error) {
	return u.repo.GetSurnameSums(ctx, text)
}

func (u *nameSearchUseCase) GetSampleNames(ctx context.Context) ([]domain.Celebrity, error) {
	return u.repo.GetSampleNames(ctx)
}
