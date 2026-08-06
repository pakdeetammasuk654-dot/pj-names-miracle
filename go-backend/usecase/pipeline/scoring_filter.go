package pipeline

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go-backend/domain"
)

// ScoringFilter ranks candidates using perfect numerology, hybrid score, and total score.
type ScoringFilter struct{}

const perfectNumerologyBonus = 8.0

func NewScoringFilter() *ScoringFilter {
	return &ScoringFilter{}
}

func (f *ScoringFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	payload := pCtx.Data
	goodMap := payload.GoodNumbers
	maxPairPoint := getMaxPairPoint(goodMap)
	var processed []domain.NameResultItem

	for _, c := range payload.Candidates {
		item := c

		if payload.Request.Mode == domain.ModeCombined {
			item.TotalSatSum = item.SatSum + payload.SurnameSatSum
			item.TotalShaSum = item.ShaSum + payload.SurnameShaSum
		} else {
			item.TotalSatSum = item.SatSum
			item.TotalShaSum = item.ShaSum
		}

		// Resolve pair meanings and total points from the central dictionary.
		satPairs, satType, satDesc, satPoint := resolveSumMeaning(item.TotalSatSum, goodMap)
		shaPairs, shaType, shaDesc, shaPoint := resolveSumMeaning(item.TotalShaSum, goodMap)

		// Ranking is an auspicious-name allowlist: every displayed pair from both
		// เลขศาสตร์ and พลังเงา must be a known D-type pair. This second check is
		// essential in combined mode because adding the surname creates new totals
		// that were not covered by the initial candidate query.
		if !allPairsAreGood(satPairs, goodMap) || !allPairsAreGood(shaPairs, goodMap) {
			continue
		}

		item.SatPairsDisplay = strings.Join(satPairs, ", ")
		item.ShaPairsDisplay = strings.Join(shaPairs, ", ")
		item.SatPairType = satType
		item.SatMiracleDesc = satDesc
		item.ShaPairType = shaType
		item.ShaMiracleDesc = shaDesc
		item.TotalScore = satPoint + shaPoint

		// Flag if both name and surname combined result in perfect D-type good numbers
		if strings.HasPrefix(satType, "D") && strings.HasPrefix(shaType, "D") {
			item.IsPerfectCombined = true
		}

		// Normalize numerology and semantic quality to the same 0-100 scale
		// before applying the user's weights.
		var semWeight float64 = 0.5
		if payload.Request.SemanticWeight != nil {
			semWeight = clamp(*payload.Request.SemanticWeight, 0, 1)
		}
		numWeight := 1.0 - semWeight

		pairCount := len(satPairs) + len(shaPairs)
		numerologyRaw := normalizeNumerologyScore(item.TotalScore, pairCount, maxPairPoint)
		semanticRaw := 0.0
		if payload.Request.WishQuery != "" {
			semanticRaw = clamp((1.0-item.VectorDistance)*100.0, 0, 100)
		} else {
			// No semantic intent means numerology owns the full score.
			numWeight = 1.0
			semWeight = 0.0
		}

		item.NumerologyScore = numerologyRaw
		item.SemanticScore = semanticRaw
		item.FinalScore = (numerologyRaw * numWeight) + (semanticRaw * semWeight)
		if item.IsPerfectCombined {
			item.PerfectBonus = perfectNumerologyBonus
			item.FinalScore = clamp(item.FinalScore+item.PerfectBonus, 0, 100)
		}
		item.RankingReasons = buildRankingReasons(item, payload.Request.DayOfBirth, payload.Request.WishQuery)

		processed = append(processed, item)
	}

	// Transparent order: hybrid score (including a disclosed perfect-pair bonus),
	// then raw numerology points, then name ID for stable ordering.
	sort.Slice(processed, func(i, j int) bool {
		if processed[i].FinalScore != processed[j].FinalScore {
			return processed[i].FinalScore > processed[j].FinalScore
		}
		if processed[i].TotalScore != processed[j].TotalScore {
			return processed[i].TotalScore > processed[j].TotalScore
		}
		return processed[i].NameID < processed[j].NameID
	})

	// Limit to top 100
	limit := payload.Request.Limit
	if limit <= 0 {
		limit = 100
	}
	if len(processed) > limit {
		processed = processed[:limit]
	}

	// Assign Rank
	for i := range processed {
		processed[i].Rank = i + 1
	}

	payload.FinalResults = processed
	return nil
}

func allPairsAreGood(pairs []string, numbers map[int]domain.NumberMeaning) bool {
	if len(pairs) == 0 {
		return false
	}

	for _, pair := range pairs {
		value, err := strconv.Atoi(pair)
		if err != nil {
			return false
		}
		meaning, ok := numbers[value]
		if !ok || !meaning.IsGood() {
			return false
		}
	}

	return true
}

func getMaxPairPoint(numbers map[int]domain.NumberMeaning) int {
	maxPoint := 0
	for _, number := range numbers {
		if number.PairPoint > maxPoint {
			maxPoint = number.PairPoint
		}
	}
	return maxPoint
}

func normalizeNumerologyScore(totalScore, pairCount, maxPairPoint int) float64 {
	if pairCount <= 0 || maxPairPoint <= 0 {
		return 0
	}
	return clamp(float64(totalScore)/float64(pairCount*maxPairPoint)*100.0, 0, 100)
}

func buildRankingReasons(item domain.NameResultItem, dayOfBirth, wishQuery string) []string {
	reasons := []string{fmt.Sprintf("คะแนนเลขศาสตร์ %.0f/100", item.NumerologyScore)}
	if item.IsPerfectCombined {
		reasons = append(reasons, fmt.Sprintf("คู่เลขมงคลสมบูรณ์แบบ (+%.0f)", item.PerfectBonus))
	}
	if wishQuery != "" && item.SemanticScore >= 70 {
		reasons = append(reasons, fmt.Sprintf("ความหมายสอดคล้อง %.0f/100", item.SemanticScore))
	}
	if strings.TrimSpace(dayOfBirth) != "" {
		reasons = append(reasons, "ผ่านการกรองกาลกิณีตามวันเกิด")
	}
	return reasons
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func resolveSumMeaning(sum int, goodMap map[int]domain.NumberMeaning) ([]string, string, string, int) {
	if sum <= 0 {
		return []string{}, "", "", 0
	}

	pairs := formatTotalValue(sum)
	var formattedPairs []string
	var types []string
	var descs []string
	totalPoint := 0

	for _, pStr := range pairs {
		val, _ := strconv.Atoi(pStr)
		twoDigitStr := fmt.Sprintf("%02d", val)
		formattedPairs = append(formattedPairs, twoDigitStr)

		if n, ok := goodMap[val]; ok {
			pType := strings.TrimSpace(n.PairType)
			if pType != "" {
				types = append(types, pType)
			}

			// Prioritize MiracleDesc; if empty fallback to MiracleDetail or PairNumber title
			desc := strings.TrimSpace(n.MiracleDesc)
			if desc == "" {
				desc = strings.TrimSpace(n.MiracleDetail)
			}
			if desc != "" {
				descs = append(descs, desc)
			}
			totalPoint += n.PairPoint
		}
	}

	typeStr := strings.Join(types, "/")
	if typeStr == "" {
		typeStr = "เลขศาสตร์มงคล"
	}

	descStr := strings.Join(descs, " + ")
	if descStr == "" {
		descStr = "พลังคู่เลขมงคลหนุนนำดวงชะตา"
	}

	return formattedPairs, typeStr, descStr, totalPoint
}
