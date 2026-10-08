package pipeline

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"go-backend/domain"
)

// ScoringFilter ranks candidates using numerology, semantic relevance, and total score.
type ScoringFilter struct{}

func NewScoringFilter() *ScoringFilter {
	return &ScoringFilter{}
}

func (f *ScoringFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	payload := pCtx.Data
	goodMap := payload.GoodNumbers
	maxPairPoint := getMaxPairPoint(goodMap)
	var processed []domain.NameResultItem

	requestedNumWeight, requestedSemWeight, requestedSoundWeight, requestedSimWeight := resolveRankingWeights(payload.Request)
	payload.RequestedWeights = domain.RankingWeights{
		Numerology: requestedNumWeight,
		Semantic:   requestedSemWeight,
		Sound:      requestedSoundWeight,
		Similarity: requestedSimWeight,
	}
	semanticAvailable := false
	semanticSource := "not_requested"
	semanticQueryModel := ""
	semanticVectorHash := ""
	candidateVectorProvenanceVerified := false
	candidateVectorProvenanceSource := "not_requested"
	if strings.TrimSpace(payload.Request.WishQuery) != "" {
		semanticSource = payload.SemanticCorpus.Source
		if semanticSource == "" {
			semanticSource = "unavailable"
		}
		candidateVectorProvenanceSource = payload.SemanticCorpus.Source
		candidateVectorProvenanceVerified = payload.SemanticCorpus.Approved
		if candidateVectorProvenanceSource == "" {
			candidateVectorProvenanceSource = "unavailable_unapproved_vector_corpus"
		}
		semanticContractAllowsScoring := payload.SemanticCorpus.Approved
		if semanticContractAllowsScoring {
			for _, candidate := range payload.Candidates {
				if semanticQueryModel == "" && candidate.SemanticQueryModel != "" {
					semanticQueryModel = candidate.SemanticQueryModel
					semanticVectorHash = candidate.SemanticVectorHash
				}
				if !semanticAvailable && candidate.SemanticSource != "" {
					semanticSource = candidate.SemanticSource
				}
				if candidate.SemanticAvailable {
					semanticAvailable = true
					semanticSource = candidate.SemanticSource
					semanticQueryModel = candidate.SemanticQueryModel
					semanticVectorHash = candidate.SemanticVectorHash
					break
				}
			}
		}
	}
	numWeight, semWeight, soundWeight, simWeight := requestedNumWeight, requestedSemWeight, requestedSoundWeight, requestedSimWeight
	if !semanticAvailable {
		semWeight = 0
		numWeight, soundWeight, simWeight = normalizeThreeWeights(numWeight, soundWeight, simWeight)
	}
	payload.AppliedWeights = domain.RankingWeights{
		Numerology: numWeight,
		Semantic:   semWeight,
		Sound:      soundWeight,
		Similarity: simWeight,
	}
	payload.SemanticAvailable = semanticAvailable
	payload.SemanticSource = semanticSource
	payload.SemanticQueryModel = semanticQueryModel
	payload.SemanticVectorHash = semanticVectorHash
	payload.CandidateVectorProvenanceVerified = candidateVectorProvenanceVerified
	payload.CandidateVectorProvenanceSource = candidateVectorProvenanceSource

	for _, c := range payload.Candidates {
		item := c

		// Keep the candidate's own pairs available for display and explanation.
		baseSatPairs := formatTotalValue(item.SatSum)
		baseShaPairs := formatTotalValue(item.ShaSum)

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

		policyPairs := append([]string{}, satPairs...)
		policyPairs = append(policyPairs, shaPairs...)
		pairPolicy := evaluateAdaptivePairPolicy(policyPairs, goodMap, requestedNumWeight)
		if !pairPolicy.Allowed {
			continue
		}

		item.SatPairsDisplay = strings.Join(satPairs, ", ")
		item.ShaPairsDisplay = strings.Join(shaPairs, ", ")
		item.SatPairs = presentPairs(satPairs, goodMap)
		item.ShaPairs = presentPairs(shaPairs, goodMap)
		item.CandidateSatPairs = presentPairs(baseSatPairs, goodMap)
		item.CandidateShaPairs = presentPairs(baseShaPairs, goodMap)
		item.SatPairType = satType
		item.SatMiracleDesc = satDesc
		item.ShaPairType = shaType
		item.ShaMiracleDesc = shaDesc
		item.HasBadPair = pairPolicy.BadCount > 0
		item.BadPairCount = pairPolicy.BadCount
		item.PairRiskLevel = pairPolicy.Level
		item.TotalScore = satPoint + shaPoint

		// This legacy flag is meaningful only in combined mode.
		if payload.Request.Mode == domain.ModeCombined {
			if strings.HasPrefix(satType, "D") && strings.HasPrefix(shaType, "D") {
				item.IsPerfectCombined = true
			}
		}

		pairCount := len(satPairs) + len(shaPairs)
		numerologyRaw := roundRankingScore(normalizeNumerologyScore(item.TotalScore, pairCount, maxPairPoint))
		semanticRaw := 0.0
		if semanticAvailable && item.SemanticAvailable {
			semanticRaw = roundRankingScore(clamp((1.0-item.VectorDistance)*100.0, 0, 100))
		}
		soundRaw := calculateSoundScore(item.EuphonyScore, item.PronunciationEase, item.RhythmScore, item.THName)
		soundCalibrated, calibrationSource := calibrateSoundScore(
			soundRaw,
			item.PhoneticVersion,
			payload.SoundCalibration,
		)
		soundCalibrated = roundRankingScore(soundCalibrated)

		similarityRaw := roundRankingScore(clamp(item.SimilarityScore, 0, 100))
		if strings.TrimSpace(payload.Request.PrototypeName) != "" {
			similarityRaw = roundRankingScore(clamp((1.0-item.SimilarityDistance)*100.0, 0, 100))
		}

		item.NumerologyScore = numerologyRaw
		item.SemanticScore = semanticRaw
		item.SimilarityScore = similarityRaw
		item.SemanticAvailable = semanticAvailable && item.SemanticAvailable
		item.SemanticModel = item.SemanticQueryModel
		if !item.SemanticAvailable || item.SemanticSource == "" {
			item.SemanticSource = semanticSource
		}
		item.CandidateVectorProvenanceVerified = candidateVectorProvenanceVerified
		item.CandidateVectorProvenanceSource = candidateVectorProvenanceSource
		item.RawSoundScore = soundRaw
		item.SoundScore = soundCalibrated
		candidateScope := candidateScope(payload.Request)
		item.SoundScope = candidateScope
		item.SoundCalibration = calibrationSource
		item.FinalScore = roundRankingScore((numerologyRaw * numWeight) +
			(semanticRaw * semWeight) +
			(soundCalibrated * soundWeight) +
			(similarityRaw * simWeight))
		item.RankingReasons = buildRankingReasons(
			item,
			payload.Request.DayOfBirth,
			payload.Request.WishQuery,
			semanticAvailable,
			candidateScope,
		)

		processed = append(processed, item)
	}

	// Transparent order: the exact user-weighted score first. Equal primary
	// scores are resolved by practical name quality; NameID is only the final
	// technical fallback and never acts as a user-facing quality signal.
	sort.Slice(processed, func(i, j int) bool {
		if processed[i].FinalScore != processed[j].FinalScore {
			return processed[i].FinalScore > processed[j].FinalScore
		}
		if processed[i].SoundScore != processed[j].SoundScore {
			return processed[i].SoundScore > processed[j].SoundScore
		}
		if numWeight > 0 && processed[i].NumerologyScore != processed[j].NumerologyScore {
			return processed[i].NumerologyScore > processed[j].NumerologyScore
		}
		if semWeight > 0 && processed[i].SemanticScore != processed[j].SemanticScore {
			return processed[i].SemanticScore > processed[j].SemanticScore
		}
		if processed[i].THName != processed[j].THName {
			return processed[i].THName < processed[j].THName
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

func presentPairs(pairs []string, numbers map[string]domain.NumberMeaning) []domain.PairPresentation {
	result := make([]domain.PairPresentation, 0, len(pairs))
	for _, value := range pairs {
		meaning, ok := numbers[value]
		result = append(result, domain.PairPresentation{
			Value:  value,
			Type:   strings.TrimSpace(meaning.PairType),
			IsGood: ok && meaning.IsGood(),
		})
	}
	return result
}

// selectAdaptiveTopResults remains a small compatibility seam for callers and
// tests. It intentionally performs no category quota: Top N is solely the
// highest N items in the already score-sorted slice.
func selectAdaptiveTopResults(sorted []domain.NameResultItem, limit int, _ float64) []domain.NameResultItem {
	if limit <= 0 || len(sorted) == 0 {
		return nil
	}
	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit]
}

func allPairsAreGood(pairs []string, numbers map[string]domain.NumberMeaning) bool {
	if len(pairs) == 0 {
		return false
	}
	for _, pair := range pairs {
		meaning, ok := numbers[pair]
		if !ok || !meaning.IsGood() {
			return false
		}
	}
	return true
}

// AdaptivePairScope is part of the response contract. All known pairs stay in
// the candidate set; their governed PairPoint contributes to numerology score.
func AdaptivePairScope(_ domain.NameSearchRequest) string { return "all_known_pairs_scored" }

type adaptivePairDecision struct {
	Allowed  bool
	BadCount int
	Level    string
}

// evaluateAdaptivePairPolicy admits every pair represented by the governed
// dictionary. Bad pair grades are a numerology signal, not a hidden fifth
// ranking rule: their PairPoint affects NumerologyScore and the UI receives a
// caution flag. Unknown pairs remain invalid because they cannot be audited.
func evaluateAdaptivePairPolicy(pairs []string, numbers map[string]domain.NumberMeaning, _ float64) adaptivePairDecision {
	decision := adaptivePairDecision{Allowed: true, Level: "strict"}
	if len(pairs) == 0 {
		decision.Allowed = false
		return decision
	}

	for _, pair := range pairs {
		meaning, ok := numbers[pair]
		if !ok {
			decision.Allowed = false
			return decision
		}
		pairType := strings.ToUpper(strings.TrimSpace(meaning.PairType))
		if strings.HasPrefix(pairType, "D") {
			continue
		}
		decision.BadCount++
		switch pairType {
		case "R5":
			decision.Level = "caution_r5"
		case "R7":
			decision.Level = "caution_r7"
		default:
			decision.Level = "caution_" + strings.ToLower(pairType)
		}
	}
	return decision
}

func getMaxPairPoint(numbers map[string]domain.NumberMeaning) int {
	maxPoint := 0
	for _, number := range numbers {
		if number.IsGood() && number.PairPoint > maxPoint {
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

func resolveRankingWeights(request domain.NameSearchRequest) (float64, float64, float64, float64) {
	// Backward compatibility for older clients that only know the original
	// semantic slider: preserve its old two-dimensional behavior.
	if request.NumerologyWeight == nil && request.SoundWeight == nil && request.SemanticWeight != nil && request.SimilarityWeight == nil {
		semantic := clamp(*request.SemanticWeight, 0, 1)
		return 1 - semantic, semantic, 0, 0
	}

	if request.NumerologyWeight == nil && request.SemanticWeight == nil && request.SoundWeight == nil && request.SimilarityWeight == nil {
		return 0.35, 0.25, 0.20, 0.20
	}

	numerology := weightValue(request.NumerologyWeight)
	semantic := weightValue(request.SemanticWeight)
	sound := weightValue(request.SoundWeight)
	similarity := weightValue(request.SimilarityWeight)
	total := numerology + semantic + sound + similarity
	if total <= 0 {
		return 0.35, 0.25, 0.20, 0.20
	}
	return numerology / total, semantic / total, sound / total, similarity / total
}

func weightValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return clamp(*value, 0, 1)
}

func normalizeTwoWeights(first, second float64) (float64, float64) {
	total := first + second
	if total <= 0 {
		return 0.5, 0.5
	}
	return first / total, second / total
}

func normalizeThreeWeights(first, second, third float64) (float64, float64, float64) {
	total := first + second + third
	if total <= 0 {
		return 0.34, 0.33, 0.33
	}
	return first / total, second / total, third / total
}

func calculateSoundScore(euphony, pronunciationEase, rhythm int, name string) float64 {
	raw := (float64(euphony) * 0.45) +
		(float64(pronunciationEase) * 0.35) +
		(float64(rhythm) * 0.20)

	length := utf8.RuneCountInString(name)
	var penalty float64 = 0
	if length < 5 {
		penalty = float64(5-length) * 2.0
	} else if length > 9 {
		penalty = float64(length-9) * 2.0
	}

	return clamp(raw-penalty, 0, 100)
}

func calibrateSoundScore(raw float64, version string, profile domain.SoundCalibrationProfile) (float64, string) {
	cleanVersion := strings.TrimSpace(version)
	if cleanVersion == "" {
		cleanVersion = domain.MissingPhoneticVersion
	}
	stats, ok := profile.Versions[cleanVersion]
	if !ok || stats.StdDev <= 0 || profile.ReferenceStdDev <= 0 || profile.ReferenceVersion == "" {
		if profile.Source != "" {
			return raw, "uncalibrated:" + profile.Source
		}
		return raw, "uncalibrated"
	}
	calibrated := profile.ReferenceMean +
		((raw-stats.Mean)/stats.StdDev)*profile.ReferenceStdDev
	if profile.Approved && profile.Source == "approved_snapshot" {
		return clamp(calibrated, 0, 100), "approved_snapshot_zscore_to_reference:" + profile.ReferenceVersion
	}
	// The fallback is deliberately labelled live: it is recomputed from the
	// current repository distribution and is not an approved immutable snapshot.
	return clamp(calibrated, 0, 100), "live_zscore_to_reference:" + profile.ReferenceVersion
}

func buildRankingReasons(item domain.NameResultItem, dayOfBirth, wishQuery string, semanticAvailable bool, scope string) []string {
	candidateLabel := "ชื่อตัว"
	if scope == string(domain.TargetSurname) {
		candidateLabel = "นามสกุล"
	}
	reasons := []string{fmt.Sprintf("คะแนนเลขศาสตร์ %.0f/100", item.NumerologyScore)}
	if item.HasBadPair {
		reasons = append(reasons, "ติดอันดับจากน้ำหนักที่เลือก แม้มีคู่เลขที่ควรพิจารณา")
	}
	if item.SoundScore > 0 {
		reasons = append(reasons, fmt.Sprintf("คุณภาพเสียง%s %.0f/100", candidateLabel, item.SoundScore))
	}
	if wishQuery != "" && semanticAvailable && item.SemanticScore >= 70 {
		reasons = append(reasons, fmt.Sprintf("ความหมายสอดคล้อง %.0f/100", item.SemanticScore))
	} else if wishQuery != "" && !semanticAvailable {
		reasons = append(reasons, "ไม่ใช้คะแนนความหมาย เนื่องจาก semantic service ไม่พร้อม")
	}
	if strings.TrimSpace(dayOfBirth) != "" {
		reasons = append(reasons, candidateLabel+"ผ่านการกรองกาลกิณีตามวันเกิด")
	}
	return reasons
}

func candidateScope(request domain.NameSearchRequest) string {
	if request.Mode == domain.ModeCombined && request.SearchTarget == domain.TargetSurname {
		return string(domain.TargetSurname)
	}
	return string(domain.TargetFirstName)
}

func clamp(value, min, max float64) float64 {
	if math.IsNaN(value) {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

const rankingDecimalScale = 1e12

// roundRankingScore makes the public precision the sorting precision too.
// Sub-display floating-point noise must not silently outrank documented tie
// breakers such as calibrated sound and Thai name text.
func roundRankingScore(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Round(value*rankingDecimalScale) / rankingDecimalScale
}

func resolveSumMeaning(sum int, goodMap map[string]domain.NumberMeaning) ([]string, string, string, int) {
	if sum <= 0 {
		return []string{}, "", "", 0
	}

	pairs := formatTotalValue(sum)
	var formattedPairs []string
	var types []string
	var descs []string
	totalPoint := 0

	for _, pStr := range pairs {
		formattedPairs = append(formattedPairs, pStr)

		if n, ok := goodMap[pStr]; ok {
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

	descStr := strings.Join(descs, " + ")
	if descStr == "" {
		// Do not invent a positive interpretation when the governed dictionary
		// has no reviewed wording for an otherwise eligible pair.
		descStr = "ยังไม่มีคำอธิบายคู่เลขในฐานข้อมูล"
	}

	return formattedPairs, typeStr, descStr, totalPoint
}
