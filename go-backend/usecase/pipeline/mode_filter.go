package pipeline

import (
	"fmt"
	"sort"
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

	// Filter good numbers map (D-type) for Step 1 candidate search
	goodMap := make(map[int]domain.NumberMeaning)
	for k, v := range allMap {
		if v.IsGood() {
			goodMap[k] = v
		}
	}

	limit := payload.Request.Limit
	if limit <= 0 {
		limit = 100
	}

	// Mode A, B & C (Semantic Wish): Fetch candidates and pass to ScoringFilter for AI Vector Ranking
	fetchLimit := 200
	var candidates []domain.NameResultItem


	goodSumList := make([]int, 0, len(goodMap))
	for num := range goodMap {
		goodSumList = append(goodSumList, num)
	}
	sort.Ints(goodSumList)

	if payload.Request.WishQuery != "" {
		var cands []domain.NameResultItem
		var err error
		
		if payload.Request.IsNameIntent {
			cands, err = f.repo.SearchSimilarNames(ctx, goodSumList, goodSumList, payload.Request.WishQuery, payload.Request.DayOfBirth, payload.Request.Gender, fetchLimit)
		} else {
			cands, err = f.repo.SearchSemanticWishNames(ctx, goodSumList, goodSumList, payload.Request.WishQuery, payload.Request.DayOfBirth, payload.Request.Gender, fetchLimit)
		}

		if err == nil && len(cands) > 0 {
			candidates = cands
		} else if err != nil {
		}
	}

	if len(candidates) == 0 {
		cands, err := f.repo.SearchCandidateNames(ctx, goodSumList, goodSumList, payload.Request.DayOfBirth, payload.Request.Gender, payload.Request.WishQuery, fetchLimit)
		if err != nil {
			return fmt.Errorf("mode filter error fetching candidates: %w", err)
		}
		candidates = cands
	}

	payload.Candidates = candidates
	return nil
}

// formatTotalValue splits multi-digit totals into 2-digit pairs (e.g. 163 -> ["16", "63"], 91 -> ["91"])
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
