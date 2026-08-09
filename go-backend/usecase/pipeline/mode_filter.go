package pipeline

import (
	"fmt"
	"strconv"

	"go-backend/domain"
	"go-backend/repository"
)

// ModeFilter implements Step 1 Filter (Combined, Standalone, or AI Semantic Wish Filter)
type ModeFilter struct {
	repo repository.NameRepository
}

func NewModeFilter(repo repository.NameRepository) *ModeFilter {
	return &ModeFilter{repo: repo}
}

func (f *ModeFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	ctx := pCtx.Ctx
	payload := pCtx.Data

	// 1. Load All Numbers Dictionary (00 - 99: D-type and R-type)
	allMap, err := f.repo.GetAllNumbersMap(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading all numbers dictionary: %w", err)
	}
	payload.GoodNumbers = allMap

	profile, err := f.repo.GetSoundCalibrationProfile(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading sound calibration: %w", err)
	}
	payload.SoundCalibration = profile
	payload.SemanticCorpus = domain.SemanticCorpusProvenance{
		Source:               "not_requested",
		EmbeddingModel:       domain.SemanticEmbeddingModel,
		InputTemplateVersion: domain.SemanticInputTemplateVersion,
		VectorDimensions:     domain.SemanticVectorDimensions,
		CorpusFilterVersion:  domain.CorpusFilterVersion,
	}

	maxSatSum, maxShaSum, err := f.repo.GetNameSumMaximums(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading name sum ranges: %w", err)
	}
	requireCombined := payload.Request.Mode == domain.ModeCombined
	targetSatSums := eligibleBaseSums(payload.SurnameSatSum, maxSatSum, allMap, requireCombined)
	targetShaSums := eligibleBaseSums(payload.SurnameShaSum, maxShaSum, allMap, requireCombined)

	// Repository limit zero means complete eligible corpus. Candidate retrieval is
	// deliberately independent of weights so a cached token can be reranked to any
	// point in the triangle without changing recall.
	var candidates []domain.NameResultItem
	if payload.Request.WishQuery == "" {
		candidates, err = f.repo.SearchCandidateNames(
			ctx, targetSatSums, targetShaSums,
			payload.Request.DayOfBirth, "", 0,
		)
	} else if payload.Request.IsNameIntent {
		payload.SemanticCorpus.Source = "not_applicable_trigram"
		candidates, err = f.repo.SearchSimilarNames(
			ctx, targetSatSums, targetShaSums,
			payload.Request.WishQuery, payload.Request.DayOfBirth, 0,
		)
	} else {
		payload.SemanticCorpus, err = f.repo.GetSemanticCorpusProvenance(ctx)
		if err != nil {
			return fmt.Errorf("mode filter error verifying semantic vector corpus: %w", err)
		}
		candidates, err = f.repo.SearchSemanticWishNames(
			ctx, targetSatSums, targetShaSums,
			payload.Request.WishQuery, payload.Request.DayOfBirth, 0,
		)
	}
	if err != nil {
		return fmt.Errorf("mode filter error fetching complete candidate corpus: %w", err)
	}
	payload.Candidates = candidates
	return nil
}

// eligibleBaseSums translates the post-surname safety rule into indexed
// sat_sum/sha_sum values that PostgreSQL can filter before ranking. It scans the
// actual sum range in names_miracle, not just the 00-99 dictionary keys.
func eligibleBaseSums(surnameSum, maxBaseSum int, numbers map[string]domain.NumberMeaning, requireCombined bool) []int {
	if maxBaseSum < 0 {
		return nil
	}

	eligible := make([]int, 0)
	for baseSum := 0; baseSum <= maxBaseSum; baseSum++ {
		if !allPairsAreGood(formatTotalValue(baseSum), numbers) {
			continue
		}
		if requireCombined && !allPairsAreGood(formatTotalValue(baseSum+surnameSum), numbers) {
			continue
		}
		eligible = append(eligible, baseSum)
	}
	return eligible
}

// formatTotalValue resolves sum-based number pairs. A single-digit total is
// read through its leading-zero pair key (6 -> 06), while authored dictionary
// keys such as "6" and "06" may still coexist for other data-owner uses.
func formatTotalValue(total int) []string {
	if total <= 0 {
		return []string{}
	}
	s := strconv.Itoa(total)
	if len(s) == 1 {
		return []string{fmt.Sprintf("%02d", total)}
	}
	if len(s) == 2 {
		return []string{s}
	}

	var pairs []string
	// Generate sliding 2-digit pairs for numbers like 163 -> 16, 63
	for i := 0; i < len(s)-1; i++ {
		pairs = append(pairs, s[i:i+2])
	}
	return pairs
}
