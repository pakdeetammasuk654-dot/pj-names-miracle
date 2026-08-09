package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase/pipeline"
)

type nameSearchUseCase struct {
	repo           repository.NameRepository
	candidateCache *candidatePoolCache
}

func NewNameSearchUseCase(repo repository.NameRepository) domain.NameSearchUseCase {
	return &nameSearchUseCase{
		repo:           repo,
		candidateCache: newCandidatePoolCache(defaultCandidateCacheCapacity, defaultCandidateCacheTTL),
	}
}

func (u *nameSearchUseCase) SearchTopNames(ctx context.Context, req domain.NameSearchRequest) ([]domain.NameResultItem, error) {
	result, err := u.SearchTopNamesWithToken(ctx, req)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (u *nameSearchUseCase) SearchTopNamesWithToken(ctx context.Context, req domain.NameSearchRequest) (domain.NameSearchResult, error) {
	var err error
	req, err = normalizeAndValidateNameSearchRequest(req)
	if err != nil {
		return domain.NameSearchResult{}, err
	}

	pool, err := u.prepareCandidatePool(ctx, req)
	if err != nil {
		return domain.NameSearchResult{}, err
	}
	outcome, err := scoreCandidatePool(ctx, pool, req)
	if err != nil {
		return domain.NameSearchResult{}, err
	}
	token, err := u.candidateCache.Put(pool)
	if err != nil {
		return domain.NameSearchResult{}, fmt.Errorf("failed to cache name candidates: %w", err)
	}

	return domain.NameSearchResult{
		CandidateToken:                    token,
		Mode:                              req.Mode,
		Data:                              outcome.data,
		EligibleCandidateCount:            len(pool.candidates),
		RequestedWeights:                  outcome.requestedWeights,
		AppliedWeights:                    outcome.appliedWeights,
		SemanticAvailable:                 outcome.semanticAvailable,
		SemanticSource:                    outcome.semanticSource,
		SemanticQueryModel:                outcome.semanticQueryModel,
		SemanticModel:                     outcome.semanticQueryModel,
		SemanticVectorHash:                outcome.semanticVectorHash,
		CandidateVectorProvenanceVerified: outcome.candidateVectorProvenanceVerified,
		CandidateVectorProvenanceSource:   outcome.candidateVectorProvenanceSource,
		SemanticCorpus:                    pool.semanticCorpus,
		Policy:                            buildNameSearchPolicy(req, pool.soundCalibration, pool.semanticCorpus),
	}, nil
}

func (u *nameSearchUseCase) RerankTopNames(ctx context.Context, req domain.NameRerankRequest) (domain.NameSearchResult, error) {
	if err := validateNameRerankRequest(req); err != nil {
		return domain.NameSearchResult{}, err
	}

	token := strings.TrimSpace(req.CandidateToken)
	pool, ok := u.candidateCache.Get(token)
	if !ok {
		return domain.NameSearchResult{}, domain.ErrCandidateTokenNotFound
	}

	rankingRequest := cloneSearchRequest(pool.request)
	rankingRequest.NumerologyWeight = cloneFloat(req.NumerologyWeight)
	rankingRequest.SemanticWeight = cloneFloat(req.SemanticWeight)
	rankingRequest.SoundWeight = cloneFloat(req.SoundWeight)
	if req.Limit > 0 {
		rankingRequest.Limit = req.Limit
	}

	outcome, err := scoreCandidatePool(ctx, pool, rankingRequest)
	if err != nil {
		return domain.NameSearchResult{}, err
	}

	return domain.NameSearchResult{
		CandidateToken:                    token,
		Mode:                              rankingRequest.Mode,
		Data:                              outcome.data,
		EligibleCandidateCount:            len(pool.candidates),
		RequestedWeights:                  outcome.requestedWeights,
		AppliedWeights:                    outcome.appliedWeights,
		SemanticAvailable:                 outcome.semanticAvailable,
		SemanticSource:                    outcome.semanticSource,
		SemanticQueryModel:                outcome.semanticQueryModel,
		SemanticModel:                     outcome.semanticQueryModel,
		SemanticVectorHash:                outcome.semanticVectorHash,
		CandidateVectorProvenanceVerified: outcome.candidateVectorProvenanceVerified,
		CandidateVectorProvenanceSource:   outcome.candidateVectorProvenanceSource,
		SemanticCorpus:                    pool.semanticCorpus,
		Policy:                            buildNameSearchPolicy(rankingRequest, pool.soundCalibration, pool.semanticCorpus),
	}, nil
}

func normalizeAndValidateNameSearchRequest(req domain.NameSearchRequest) (domain.NameSearchRequest, error) {
	if req.Limit == 0 {
		req.Limit = 100
	} else if req.Limit < 0 || req.Limit > 100 {
		return domain.NameSearchRequest{}, fmt.Errorf("%w: limit must be between 1 and 100", domain.ErrInvalidSearchRequest)
	}
	if req.Mode == "" {
		req.Mode = domain.ModeStandalone
	}
	req.Surname = strings.TrimSpace(req.Surname)
	req.DayOfBirth = strings.ToLower(strings.TrimSpace(req.DayOfBirth))
	req.WishQuery = strings.TrimSpace(req.WishQuery)

	switch req.Mode {
	case domain.ModeStandalone, domain.ModeCombined, domain.ModeSemantic:
	default:
		return domain.NameSearchRequest{}, fmt.Errorf("%w: unsupported mode %q", domain.ErrInvalidSearchRequest, req.Mode)
	}
	if req.Mode == domain.ModeCombined {
		if req.Surname == "" {
			return domain.NameSearchRequest{}, fmt.Errorf("%w: surname is required in combined mode", domain.ErrInvalidSearchRequest)
		}
	} else {
		// Prevent stale form state from leaking into standalone result metadata.
		req.Surname = ""
	}
	if req.Mode == domain.ModeSemantic && req.WishQuery == "" {
		return domain.NameSearchRequest{}, fmt.Errorf("%w: wish_query is required in semantic mode", domain.ErrInvalidSearchRequest)
	}
	if req.IsNameIntent && req.WishQuery == "" {
		return domain.NameSearchRequest{}, fmt.Errorf("%w: wish_query is required for name intent", domain.ErrInvalidSearchRequest)
	}
	if req.DayOfBirth != "" && !isValidDayOfBirth(req.DayOfBirth) {
		return domain.NameSearchRequest{}, fmt.Errorf("%w: unsupported day_of_birth %q", domain.ErrInvalidSearchRequest, req.DayOfBirth)
	}
	if err := validateInitialWeights(req); err != nil {
		return domain.NameSearchRequest{}, err
	}
	return req, nil
}

func isValidDayOfBirth(day string) bool {
	switch day {
	case "sunday", "monday", "tuesday", "wednesday1", "wednesday2", "thursday", "friday", "saturday":
		return true
	default:
		return false
	}
}

func validateInitialWeights(req domain.NameSearchRequest) error {
	present := 0
	for _, weight := range []*float64{req.NumerologyWeight, req.SemanticWeight, req.SoundWeight} {
		if weight != nil {
			present++
		}
	}
	if present == 0 {
		return nil
	}
	// Preserve the legacy semantic-only slider contract.
	legacySemanticOnly := present == 1 && req.SemanticWeight != nil
	if present != 3 && !legacySemanticOnly {
		return fmt.Errorf("%w: provide all three weights or semantic_weight alone for the legacy contract", domain.ErrInvalidSearchRequest)
	}
	total := 0.0
	for name, weight := range map[string]*float64{
		"numerology_weight": req.NumerologyWeight,
		"semantic_weight":   req.SemanticWeight,
		"sound_weight":      req.SoundWeight,
	} {
		if weight == nil {
			continue
		}
		if math.IsNaN(*weight) || math.IsInf(*weight, 0) || *weight < 0 || *weight > 1 {
			return fmt.Errorf("%w: %s must be between 0 and 1", domain.ErrInvalidSearchRequest, name)
		}
		total += *weight
	}
	if total <= 0 && !legacySemanticOnly {
		return fmt.Errorf("%w: at least one weight must be greater than zero", domain.ErrInvalidSearchRequest)
	}
	return nil
}

func (u *nameSearchUseCase) prepareCandidatePool(ctx context.Context, req domain.NameSearchRequest) (cachedCandidatePool, error) {
	payload := &domain.NameSearchPayload{Request: req}
	if req.Mode == domain.ModeCombined && strings.TrimSpace(req.Surname) != "" {
		satSum, shaSum, err := u.repo.GetSurnameSums(ctx, req.Surname)
		if err != nil {
			if errors.Is(err, domain.ErrUnmappedCharacters) {
				return cachedCandidatePool{}, fmt.Errorf("%w: invalid surname: %v", domain.ErrInvalidSearchRequest, err)
			}
			return cachedCandidatePool{}, err
		}
		payload.SurnameSatSum = satSum
		payload.SurnameShaSum = shaSum
	}

	pCtx := &domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  ctx,
		Data: payload,
	}

	// Prepare every value ScoringFilter needs. This is the only phase allowed to
	// call the repository; reranks resume from the cached output of this phase.
	pipe := domain.NewPipeline[*domain.NameSearchPayload](
		pipeline.NewModeFilter(u.repo), // Step 1: Filter D10/D8/D5 by Search Mode
		pipeline.NewKalakiniFilter(),   // Step 2: Day of Birth Kalakini Safety Filter (100% Kalakini Free)
	)

	if err := pipe.Execute(pCtx); err != nil {
		return cachedCandidatePool{}, err
	}

	return cachedCandidatePool{
		request:          cloneSearchRequest(pCtx.Data.Request),
		surnameSatSum:    pCtx.Data.SurnameSatSum,
		surnameShaSum:    pCtx.Data.SurnameShaSum,
		goodNumbers:      pCtx.Data.GoodNumbers,
		soundCalibration: pCtx.Data.SoundCalibration,
		semanticCorpus:   pCtx.Data.SemanticCorpus,
		candidates:       pCtx.Data.Candidates,
	}, nil
}

type scoringOutcome struct {
	data                              []domain.NameResultItem
	requestedWeights                  domain.RankingWeights
	appliedWeights                    domain.RankingWeights
	semanticAvailable                 bool
	semanticSource                    string
	semanticQueryModel                string
	semanticVectorHash                string
	candidateVectorProvenanceVerified bool
	candidateVectorProvenanceSource   string
}

func scoreCandidatePool(ctx context.Context, pool cachedCandidatePool, req domain.NameSearchRequest) (scoringOutcome, error) {
	payload := &domain.NameSearchPayload{
		Request:          cloneSearchRequest(req),
		SurnameSatSum:    pool.surnameSatSum,
		SurnameShaSum:    pool.surnameShaSum,
		GoodNumbers:      pool.goodNumbers,
		SoundCalibration: cloneSoundCalibrationProfile(pool.soundCalibration),
		SemanticCorpus:   pool.semanticCorpus,
		Candidates:       cloneNameResults(pool.candidates),
	}
	pCtx := &domain.PipelineContext[*domain.NameSearchPayload]{Ctx: ctx, Data: payload}
	if err := pipeline.NewScoringFilter().Execute(pCtx); err != nil {
		return scoringOutcome{}, err
	}
	return scoringOutcome{
		data:                              pCtx.Data.FinalResults,
		requestedWeights:                  pCtx.Data.RequestedWeights,
		appliedWeights:                    pCtx.Data.AppliedWeights,
		semanticAvailable:                 pCtx.Data.SemanticAvailable,
		semanticSource:                    pCtx.Data.SemanticSource,
		semanticQueryModel:                pCtx.Data.SemanticQueryModel,
		semanticVectorHash:                pCtx.Data.SemanticVectorHash,
		candidateVectorProvenanceVerified: pCtx.Data.CandidateVectorProvenanceVerified,
		candidateVectorProvenanceSource:   pCtx.Data.CandidateVectorProvenanceSource,
	}, nil
}

func buildNameSearchPolicy(
	req domain.NameSearchRequest,
	calibration domain.SoundCalibrationProfile,
	semanticCorpus domain.SemanticCorpusProvenance,
) domain.NameSearchPolicy {
	pairScope := "first_name_all_d_pairs"
	if req.Mode == domain.ModeCombined {
		pairScope = "first_name_and_combined_all_d_pairs"
	}
	kalakiniScope := "not_applied"
	if req.DayOfBirth != "" {
		kalakiniScope = "first_name"
	}
	return domain.NameSearchPolicy{
		PairScope:                     pairScope,
		KalakiniScope:                 kalakiniScope,
		SoundScope:                    "first_name",
		NumerologyRuleVersion:         domain.NumerologyRuleVersion,
		SoundFormulaVersion:           domain.SoundFormulaVersion,
		RankingPrecisionVersion:       domain.RankingPrecisionVersion,
		SoundCalibrationSource:        calibration.Source,
		SoundCalibrationReference:     calibration.ReferenceVersion,
		SoundCalibrationApproved:      calibration.Approved,
		SoundCalibrationAuditAt:       calibration.SourceAuditAt,
		SoundProfileLoadedAt:          calibration.LoadedAt,
		SoundEvaluationVersion:        calibration.EvaluationVersion,
		SoundCorpusFingerprint:        calibration.CorpusFingerprintSHA256,
		CorpusFilterVersion:           domain.CorpusFilterVersion,
		SemanticQueryModel:            domain.SemanticEmbeddingModel,
		SemanticInputTemplate:         domain.SemanticInputTemplateVersion,
		SemanticVectorDimensions:      domain.SemanticVectorDimensions,
		SemanticCorpusApproved:        semanticCorpus.Approved,
		SemanticCorpusSource:          semanticCorpus.Source,
		SemanticCorpusFingerprint:     semanticCorpus.CorpusVectorSHA256,
		SemanticMembershipFingerprint: semanticCorpus.CorpusMembershipSHA256,
		SemanticCorpusAuditAt:         semanticCorpus.SourceAuditAt,
		SemanticEvaluationVersion:     semanticCorpus.EvaluationVersion,
	}
}

func validateNameRerankRequest(req domain.NameRerankRequest) error {
	token := strings.TrimSpace(req.CandidateToken)
	decodedToken, err := base64.RawURLEncoding.DecodeString(token)
	if token == "" || err != nil || len(decodedToken) != candidateTokenBytes {
		return fmt.Errorf("%w: candidate_token must be a valid search token", domain.ErrInvalidRerankRequest)
	}

	weights := []struct {
		name  string
		value *float64
	}{
		{name: "numerology_weight", value: req.NumerologyWeight},
		{name: "semantic_weight", value: req.SemanticWeight},
		{name: "sound_weight", value: req.SoundWeight},
	}
	total := 0.0
	for _, weight := range weights {
		if weight.value == nil {
			return fmt.Errorf("%w: %s is required", domain.ErrInvalidRerankRequest, weight.name)
		}
		value := *weight.value
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return fmt.Errorf("%w: %s must be between 0 and 1", domain.ErrInvalidRerankRequest, weight.name)
		}
		total += value
	}
	if total <= 0 {
		return fmt.Errorf("%w: at least one weight must be greater than zero", domain.ErrInvalidRerankRequest)
	}
	if req.Limit < 0 || req.Limit > 100 {
		return fmt.Errorf("%w: limit must be between 1 and 100", domain.ErrInvalidRerankRequest)
	}
	return nil
}

func (u *nameSearchUseCase) GetAllNumbersMap(ctx context.Context) (map[string]domain.NumberMeaning, error) {
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
