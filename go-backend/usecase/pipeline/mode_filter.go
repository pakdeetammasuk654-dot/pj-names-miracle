package pipeline

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
	totalStart := time.Now()
	defer func() {
		log.Printf("[Perf][ModeFilter] total=%s candidates=%d",
			time.Since(totalStart).Round(time.Millisecond),
			len(pCtx.Data.Candidates),
		)
	}()

	ctx := pCtx.Ctx
	payload := pCtx.Data

	// 1. Load All Numbers Dictionary (00 - 99: D-type and R-type)
	stepStart := time.Now()
	allMap, err := f.repo.GetAllNumbersMap(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading all numbers dictionary: %w", err)
	}
	log.Printf("[Perf][ModeFilter] GetAllNumbersMap duration=%s rows=%d",
		time.Since(stepStart).Round(time.Millisecond),
		len(allMap),
	)
	payload.GoodNumbers = allMap

	stepStart = time.Now()
	profile, err := f.repo.GetSoundCalibrationProfile(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading sound calibration: %w", err)
	}
	log.Printf("[Perf][ModeFilter] GetSoundCalibrationProfile duration=%s source=%s approved=%t reference_version=%s versions=%d",
		time.Since(stepStart).Round(time.Millisecond),
		profile.Source,
		profile.Approved,
		profile.ReferenceVersion,
		len(profile.Versions),
	)
	payload.SoundCalibration = profile
	payload.SemanticCorpus = domain.SemanticCorpusProvenance{
		Source:               "not_requested",
		EmbeddingModel:       domain.SemanticEmbeddingModel,
		InputTemplateVersion: domain.SemanticInputTemplateVersion,
		VectorDimensions:     domain.SemanticVectorDimensions,
		CorpusFilterVersion:  domain.CorpusFilterVersion,
	}

	stepStart = time.Now()
	maxSatSum, maxShaSum, err := f.repo.GetNameSumMaximums(ctx)
	if err != nil {
		return fmt.Errorf("mode filter error loading name sum ranges: %w", err)
	}
	log.Printf("[Perf][ModeFilter] GetNameSumMaximums duration=%s max_sat=%d max_sha=%d",
		time.Since(stepStart).Round(time.Millisecond),
		maxSatSum,
		maxShaSum,
	)
	requireCombined := payload.Request.Mode == domain.ModeCombined
	stepStart = time.Now()
	targetSatSums := eligibleBaseSums(payload.SurnameSatSum, maxSatSum, allMap, requireCombined)
	targetShaSums := eligibleBaseSums(payload.SurnameShaSum, maxShaSum, allMap, requireCombined)
	log.Printf("[Perf][ModeFilter] eligibleBaseSums duration=%s target_sat_count=%d target_sha_count=%d combined=%t",
		time.Since(stepStart).Round(time.Millisecond),
		len(targetSatSums),
		len(targetShaSums),
		requireCombined,
	)

	// Repository limit zero means complete eligible corpus. Candidate retrieval is
	// deliberately independent of weights so a cached token can be reranked to any
	// point in the triangle without changing recall.
	var candidates []domain.NameResultItem
	stepStart = time.Now()
	retrievalKind := "candidate"
	if payload.Request.IsNameIntent {
		retrievalKind = "semantic_and_trigram"
		seedName := strings.TrimSpace(payload.Request.PrototypeName)

		if payload.Request.WishQuery == "" {
			payload.SemanticCorpus.Source = "not_requested"
			candidates, err = f.repo.SearchSimilarNames(
				ctx, targetSatSums, targetShaSums,
				seedName, payload.Request.DayOfBirth, 0,
			)
		} else {
			provenanceStart := time.Now()
			payload.SemanticCorpus, err = f.repo.GetSemanticCorpusProvenance(ctx)
			if err != nil {
				return fmt.Errorf("mode filter error verifying semantic vector corpus: %w", err)
			}
			log.Printf("[Perf][ModeFilter] GetSemanticCorpusProvenance duration=%s", time.Since(provenanceStart).Round(time.Millisecond))
			candidates, err = f.repo.SearchSemanticAndSimilarNames(
				ctx, targetSatSums, targetShaSums,
				seedName, payload.Request.WishQuery, payload.Request.DayOfBirth, 0,
			)
		}
	} else if payload.Request.WishQuery == "" {
		candidates, err = f.repo.SearchCandidateNames(
			ctx, targetSatSums, targetShaSums,
			payload.Request.DayOfBirth, "", 0,
		)
	} else {
		retrievalKind = "semantic"
		provenanceStart := time.Now()
		payload.SemanticCorpus, err = f.repo.GetSemanticCorpusProvenance(ctx)
		if err != nil {
			return fmt.Errorf("mode filter error verifying semantic vector corpus: %w", err)
		}
		log.Printf("[Perf][ModeFilter] GetSemanticCorpusProvenance duration=%s source=%s approved=%t",
			time.Since(provenanceStart).Round(time.Millisecond),
			payload.SemanticCorpus.Source,
			payload.SemanticCorpus.Approved,
		)
		candidates, err = f.repo.SearchSemanticWishNames(
			ctx, targetSatSums, targetShaSums,
			payload.Request.WishQuery, payload.Request.DayOfBirth, 0,
		)
	}
	if err != nil {
		return fmt.Errorf("mode filter error fetching complete candidate corpus: %w", err)
	}
	log.Printf("[Perf][ModeFilter] %s retrieval duration=%s candidates=%d day=%s wish=%t",
		retrievalKind,
		time.Since(stepStart).Round(time.Millisecond),
		len(candidates),
		payload.Request.DayOfBirth,
		payload.Request.WishQuery != "",
	)
	payload.Candidates = candidates
	return nil
}

// eligibleBaseSums builds a weight-independent candidate superset. Reranking
// must be able to reveal a narrowly controlled bad pair after the user lowers
// the numerology weight, without another database or embedding request.
// ScoringFilter applies the exact weight-dependent policy to this cached pool.
func eligibleBaseSums(surnameSum, maxBaseSum int, numbers map[string]domain.NumberMeaning, requireCombined bool) []int {
	if maxBaseSum < 0 {
		return nil
	}

	eligible := make([]int, 0)
	for baseSum := 0; baseSum <= maxBaseSum; baseSum++ {
		basePairs := formatTotalValue(baseSum)
		if requireCombined {
			if !allPairsAreKnown(basePairs, numbers) ||
				!allPairsAreKnown(formatTotalValue(baseSum+surnameSum), numbers) {
				continue
			}
		} else if !allPairsAreKnown(basePairs, numbers) {
			continue
		}
		eligible = append(eligible, baseSum)
	}
	return eligible
}

// The cache must cover the 0% numerology posture, where pair grades do not
// reject a candidate. Unknown pairs remain invalid because they cannot be
// explained or rendered from the governed dictionary.
func allPairsAreKnown(pairs []string, numbers map[string]domain.NumberMeaning) bool {
	if len(pairs) == 0 {
		return false
	}
	for _, pair := range pairs {
		_, ok := numbers[pair]
		if !ok {
			return false
		}
	}
	return true
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
