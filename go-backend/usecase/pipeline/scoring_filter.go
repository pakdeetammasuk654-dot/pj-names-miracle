package pipeline

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go-backend/domain"
)

// ScoringFilter ranks candidate names prioritizing Chosen Taksa Target #1, AI Semantic Vector Match #2, Total Score #3
type ScoringFilter struct{}

func NewScoringFilter() *ScoringFilter {
	return &ScoringFilter{}
}

func (f *ScoringFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	payload := pCtx.Data
	goodMap := payload.GoodNumbers
	dayOfBirth := payload.Request.DayOfBirth
	taksaTarget := payload.Request.TaksaTarget
	if taksaTarget == "" {
		taksaTarget = "sri"
	}

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

		// Priority #1: Calculate user-chosen Taksa target character count (Sri, Dech, Montri, Mula, Utsaha, Ayu, Bariwan)
		item.TaksaCount = CountTaksaRunes(item.THName, dayOfBirth, taksaTarget)
		item.TaksaLabel = GetTaksaLabelThai(taksaTarget)
		item.TaksaChars = GetTaksaMatchedChars(item.THName, dayOfBirth, taksaTarget)

		// Priority #3: Resolve PairType meanings & total points
		satPairs, satType, satDesc, satPoint := resolveSumMeaning(item.TotalSatSum, goodMap)
		shaPairs, shaType, shaDesc, shaPoint := resolveSumMeaning(item.TotalShaSum, goodMap)

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

		// Calculate Hybrid FinalScore
		var semWeight float64 = 0.5
		if payload.Request.SemanticWeight != nil {
			semWeight = *payload.Request.SemanticWeight
		}
		numWeight := 1.0 - semWeight

		numerologyRaw := float64(item.TotalScore) / 2.0
		semanticRaw := 0.0
		if payload.Request.WishQuery != "" {
			semanticRaw = (1.0 - item.VectorDistance) * 100.0
			if semanticRaw < 0 {
				semanticRaw = 0
			}
		} else {
			// If no wish query, numerology dictates 100% of the score implicitly, but we just use numWeight
			semanticRaw = 0
			// Force numWeight to 1.0 if no semantic wish
			numWeight = 1.0
			semWeight = 0.0
		}
		
		item.FinalScore = (numerologyRaw * numWeight) + (semanticRaw * semWeight)

		processed = append(processed, item)
	}

	// Smart Hybrid Ranking Strategy:
	// Priority #1: IsPerfectCombined (Descending) - Names where (Name + Surname) equals perfect D-type pair numbers float to top
	// Priority #2: FinalScore (Descending) - Hybrid weighted score (Numerology vs Semantic)

	sort.Slice(processed, func(i, j int) bool {
		if processed[i].IsPerfectCombined != processed[j].IsPerfectCombined {
			return processed[i].IsPerfectCombined
		}
		if processed[i].FinalScore != processed[j].FinalScore {
			return processed[i].FinalScore > processed[j].FinalScore
		}
		return processed[i].TotalScore > processed[j].TotalScore
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

// countThaiSyllables calculates actual Thai syllable count accurately
func countThaiSyllables(name string) int {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return 0
	}

	// Names with <= 2 runes (e.g. เช, ชู, พล, นพ, ชา) are strictly 1 syllable in names
	if len(runes) <= 2 {
		return 1
	}

	// Count explicit Thai vowel clusters/nuclei
	nucleusCount := 0
	for _, r := range runes {
		switch r {
		case 'ะ', 'า', 'ำ', 'ิ', 'ี', 'ึ', 'ื', 'ุ', 'ู', 'เ', 'แ', 'โ', 'ใ', 'ไ', 'ฤ':
			nucleusCount++
		}
	}

	if nucleusCount == 0 {
		nucleusCount = 1
	}

	return nucleusCount
}
