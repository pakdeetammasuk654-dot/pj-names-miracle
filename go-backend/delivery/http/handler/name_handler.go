package handler

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
)

const freeVisibleNameCount = 3

type NameHandler struct {
	useCase       domain.NameSearchUseCase
	dimensionRepo repository.DimensionLabelRepository
	entitlement   usecase.EntitlementService
}

func (h *NameHandler) SetEntitlementService(service usecase.EntitlementService) {
	h.entitlement = service
}

func (h *NameHandler) applyPremiumVisibility(c *gin.Context, data []domain.NameResultItem) ([]domain.NameResultItem, bool) {
	if data == nil {
		data = []domain.NameResultItem{}
	}
	// Unit tests and embedders that do not configure billing retain the existing
	// handler contract. Production always injects the entitlement service.
	if h.entitlement == nil {
		return data, true
	}
	unlocked := h.entitlement.Authorize(c.GetHeader("X-Entitlement-Token"))
	if unlocked || len(data) <= freeVisibleNameCount {
		return data, unlocked
	}
	redacted := append([]domain.NameResultItem(nil), data...)
	for i := freeVisibleNameCount; i < len(redacted); i++ {
		redacted[i].THName = maskThaiName(redacted[i].THName)
		redacted[i].Meaning = ""
		redacted[i].RankingReasons = nil
	}
	return redacted, false
}

func maskThaiName(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return "***"
	}
	visible := 1
	if len(runes) > 4 {
		visible = 2
	}
	return string(runes[:visible]) + "***"
}

func NewNameHandler(useCase domain.NameSearchUseCase, dimensionRepo repository.DimensionLabelRepository) *NameHandler {
	return &NameHandler{useCase: useCase, dimensionRepo: dimensionRepo}
}

func (h *NameHandler) GetRandomMeanings(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "8"))
	if err != nil || limit < 1 || limit > 108 {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "limit must be between 1 and 108"})
		return
	}
	meanings, err := h.useCase.GetRandomMeanings(c.Request.Context(), limit)
	if err != nil {
		log.Printf("random meanings failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to load random meanings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":    false,
		"meanings": meanings,
		"count":    len(meanings),
	})
}

func (h *NameHandler) SearchTopNames(c *gin.Context) {
	var req domain.NameSearchRequest
	if c.Request.Method == http.MethodGet {
		req.Mode = domain.SearchMode(c.DefaultQuery("mode", "standalone"))
		req.Surname = c.Query("surname")
		req.DayOfBirth = c.Query("day_of_birth")
		req.WishQuery = c.Query("wish_query")
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Invalid request payload: " + err.Error(),
			})
			return
		}
	}

	result, err := h.useCase.SearchTopNamesWithToken(c.Request.Context(), req)
	if err != nil {
		log.Printf("name search failed: %v", err)
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidSearchRequest) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	visibleData, premiumUnlocked := h.applyPremiumVisibility(c, result.Data)
	c.JSON(http.StatusOK, gin.H{
		"error":                                false,
		"mode":                                 result.Mode,
		"applied_mode":                         result.Mode,
		"count":                                len(result.Data),
		"returned_count":                       len(result.Data),
		"eligible_candidate_count":             result.EligibleCandidateCount,
		"data":                                 visibleData,
		"premium_unlocked":                     premiumUnlocked,
		"free_visible_count":                   freeVisibleNameCount,
		"candidate_token":                      result.CandidateToken,
		"requested_weights":                    result.RequestedWeights,
		"applied_weights":                      result.AppliedWeights,
		"semantic_available":                   result.SemanticAvailable,
		"semantic_source":                      result.SemanticSource,
		"semantic_model":                       result.SemanticQueryModel,
		"semantic_query_model":                 result.SemanticQueryModel,
		"semantic_vector_hash":                 result.SemanticVectorHash,
		"candidate_vector_provenance_verified": result.CandidateVectorProvenanceVerified,
		"candidate_vector_provenance_source":   result.CandidateVectorProvenanceSource,
		"policy":                               result.Policy,
		"message":                              fmt.Sprintf("Successfully fetched %d auspicious names", len(result.Data)),
	})
}

// SuggestNames is the lightweight type-ahead endpoint for the reference-name
// field. It deliberately runs without a birthday: suggestions must first be
// good in both เลขศาสตร์ and พลังเงา; Kalakini is an optional second step.
func (h *NameHandler) SuggestNames(c *gin.Context) {
	query := strings.TrimSpace(c.Query("query"))
	if utf8.RuneCountInString(query) < 2 {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"error": false, "data": []domain.NameResultItem{}, "preview_pairs": []suggestionPairPreview{}, "count": 0, "premium_unlocked": false})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "8"))
	if err != nil || limit < 1 || limit > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "limit must be between 1 and 12"})
		return
	}
	results, err := h.useCase.SuggestNames(c.Request.Context(), query, limit)
	if err != nil {
		log.Printf("name suggestions failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to suggest names"})
		return
	}
	unlocked := h.entitlement != nil && h.entitlement.AuthorizeVerified(c.GetHeader("X-Entitlement-Token"))
	c.Header("Cache-Control", "no-store")
	if !unlocked {
		// The teaser contains partially masked names and audited pair data.
		// Never serialize real full names, IDs, meanings, or candidate tokens.
		previews := make([]suggestionPairPreview, 0, min(len(results), 2))
		for _, item := range results {
			if len(previews) == 2 {
				break
			}
			previews = append(previews, suggestionPairPreview{
				MaskedName: maskSuggestionName(item.THName, query),
				SatPairs:   item.CandidateSatPairs,
				ShaPairs:   item.CandidateShaPairs,
			})
		}
		c.JSON(http.StatusOK, gin.H{"error": false, "data": []domain.NameResultItem{}, "preview_pairs": previews, "count": len(results), "premium_unlocked": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"error": false, "data": results, "preview_pairs": []suggestionPairPreview{}, "count": len(results), "premium_unlocked": true})
}

func maskSuggestionName(name, query string) string {
	runes := []rune(strings.TrimSpace(name))
	n := len(runes)
	if n == 0 {
		return "***"
	}
	visible := 1
	if n > 4 {
		visible = 2
	}
	qRunes := len([]rune(strings.TrimSpace(query)))
	if qRunes > visible && strings.HasPrefix(strings.TrimSpace(name), strings.TrimSpace(query)) {
		visible = qRunes
	}
	if visible >= n {
		visible = max(1, n-1)
	}
	return string(runes[:visible]) + "***"
}

type suggestionPairPreview struct {
	MaskedName string                    `json:"masked_name"`
	SatPairs   []domain.PairPresentation `json:"sat_pairs"`
	ShaPairs   []domain.PairPresentation `json:"sha_pairs"`
}

// RerankTopNames reuses the repository-independent candidate pool created by
// SearchTopNames. This makes slider interactions fast and prevents a rerank
// from triggering another database or embedding request.
func (h *NameHandler) RerankTopNames(c *gin.Context) {
	var req domain.NameRerankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request payload: " + err.Error(),
		})
		return
	}

	result, err := h.useCase.RerankTopNames(c.Request.Context(), req)
	if err != nil {
		log.Printf("name rerank failed: %v", err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrInvalidRerankRequest):
			status = http.StatusBadRequest
		case errors.Is(err, domain.ErrCandidateTokenNotFound):
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	visibleData, premiumUnlocked := h.applyPremiumVisibility(c, result.Data)
	c.JSON(http.StatusOK, gin.H{
		"error":                                false,
		"mode":                                 result.Mode,
		"applied_mode":                         result.Mode,
		"count":                                len(result.Data),
		"returned_count":                       len(result.Data),
		"eligible_candidate_count":             result.EligibleCandidateCount,
		"data":                                 visibleData,
		"premium_unlocked":                     premiumUnlocked,
		"free_visible_count":                   freeVisibleNameCount,
		"candidate_token":                      result.CandidateToken,
		"requested_weights":                    result.RequestedWeights,
		"applied_weights":                      result.AppliedWeights,
		"semantic_available":                   result.SemanticAvailable,
		"semantic_source":                      result.SemanticSource,
		"semantic_model":                       result.SemanticQueryModel,
		"semantic_query_model":                 result.SemanticQueryModel,
		"semantic_vector_hash":                 result.SemanticVectorHash,
		"candidate_vector_provenance_verified": result.CandidateVectorProvenanceVerified,
		"candidate_vector_provenance_source":   result.CandidateVectorProvenanceSource,
		"policy":                               result.Policy,
		"message":                              "Successfully reranked auspicious names",
	})
}

// GetCelebrities returns a list of celebrity names for the UI playground / naming examples
func (h *NameHandler) GetCelebrities(c *gin.Context) {
	celebrities, err := h.useCase.GetSampleNames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch sample names",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  celebrities,
	})
}

func formatTotalValueForDecode(total int) []string {
	if total <= 0 {
		return []string{}
	}
	totalStr := fmt.Sprintf("%d", total)
	if len(totalStr) == 1 {
		return []string{fmt.Sprintf("%02d", total)}
	}
	if len(totalStr) == 2 {
		return []string{totalStr}
	}
	var pairs []string
	for i := 0; i < len(totalStr)-1; i++ {
		pairs = append(pairs, totalStr[i:i+2])
	}
	return pairs
}

func decodePairMetadata(pairs []string, numbers map[string]domain.NumberMeaning) ([]string, []int, []bool, []string) {
	pairTypes := make([]string, 0, len(pairs))
	pairPoints := make([]int, 0, len(pairs))
	pairIsGood := make([]bool, 0, len(pairs))
	pairDescs := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		meaning, ok := numbers[pair]
		if !ok {
			pairTypes = append(pairTypes, "")
			pairPoints = append(pairPoints, 0)
			pairIsGood = append(pairIsGood, false)
			pairDescs = append(pairDescs, "")
			continue
		}
		pairTypes = append(pairTypes, meaning.PairType)
		pairPoints = append(pairPoints, meaning.PairPoint)
		pairIsGood = append(pairIsGood, meaning.IsGood())
		pairDescs = append(pairDescs, strings.TrimSpace(meaning.MiracleDesc))
	}
	return pairTypes, pairPoints, pairIsGood, pairDescs
}

func (h *NameHandler) DecodeName(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	dayOfBirth := strings.ToLower(strings.TrimSpace(c.Query("day_of_birth")))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "name is required"})
		return
	}

	words := strings.Fields(name)
	var satSum, shaSum int
	var combinedMeaning string

	for _, word := range words {
		wSat, wSha, wMeaning, err := h.useCase.GetSumsByTHName(c.Request.Context(), word)
		if err != nil {
			wSat, wSha, err = h.useCase.GetSumsForText(c.Request.Context(), word)
			if err != nil {
				status := http.StatusInternalServerError
				message := "Failed to decode name"
				if errors.Is(err, domain.ErrUnmappedCharacters) {
					status = http.StatusBadRequest
					message = err.Error()
				}
				c.JSON(status, gin.H{"error": true, "message": message})
				return
			}
			wMeaning = ""
		}
		satSum += wSat
		shaSum += wSha
		if wMeaning != "" {
			combinedMeaning = domain.CombineMeanings(combinedMeaning, wMeaning)
		}
	}

	allNumsMap, err := h.useCase.GetAllNumbersMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to load number dictionary"})
		return
	}

	pairs := append(formatTotalValueForDecode(satSum), formatTotalValueForDecode(shaSum)...)
	pairTypes, pairPoints, pairIsGood, pairDescs := decodePairMetadata(pairs, allNumsMap)

	// Calculate 4D Name Profile Metrics with high sensitivity and dynamic range
	// 1. Numerology Score: based on pair points, good ratio, and total points
	var totalPts float64
	var goodCount int
	for i, pt := range pairPoints {
		totalPts += float64(pt)
		if i < len(pairIsGood) && pairIsGood[i] {
			goodCount++
		}
	}
	numScore := 40.0
	if len(pairPoints) > 0 {
		avgPt := totalPts / float64(len(pairPoints))
		// Base scale from pair points (-20 to +20) mapped to 20..95
		numScore = 55.0 + (avgPt * 1.8)
		if goodCount == len(pairPoints) && goodCount > 0 {
			numScore += 10.0
		} else if goodCount == 0 {
			numScore -= 15.0
		}
		numScore = math.Max(15.0, math.Min(98.0, numScore))
	}

	// 2. Meaning Score: based on depth/length of Pali/Sanskrit dictionary meaning
	meaningScore := 35.0
	cleanMeaning := strings.TrimSpace(combinedMeaning)
	if cleanMeaning != "" {
		runeLen := utf8.RuneCountInString(cleanMeaning)
		if runeLen >= 30 {
			meaningScore = 95.0
		} else if runeLen >= 15 {
			meaningScore = 80.0 + float64(runeLen-15)
		} else {
			meaningScore = 55.0 + float64(runeLen*2)
		}
	}

	// 3. Sound Score: phonetics balance, syllable flow, smooth vowels
	cleanName := strings.ReplaceAll(name, " ", "")
	charLen := utf8.RuneCountInString(cleanName)
	soundScore := 50.0
	switch {
	case charLen >= 2 && charLen <= 4:
		soundScore = 92.0
	case charLen == 5:
		soundScore = 82.0
	case charLen == 6:
		soundScore = 70.0
	default:
		soundScore = math.Max(25.0, 60.0-float64(charLen*3))
	}

	// 4. Resonance / Morphological Identity: consonant-vowel rhythm & root morphology
	resonanceScore := 50.0
	if cleanMeaning != "" {
		resonanceScore += 25.0
	}
	if len(pairPoints) > 0 && goodCount > 0 {
		resonanceScore += float64(goodCount) / float64(len(pairPoints)) * 20.0
	}
	// Bonus for classic auspicious prefixes
	if strings.HasPrefix(cleanName, "กัญ") || strings.HasPrefix(cleanName, "ณั") ||
		strings.HasPrefix(cleanName, "ปั") || strings.HasPrefix(cleanName, "ธน") ||
		strings.HasPrefix(cleanName, "พร") || strings.HasPrefix(cleanName, "ศิ") {
		resonanceScore += 5.0
	}
	resonanceScore = math.Max(20.0, math.Min(96.0, resonanceScore))

	meaningLabel := "ความหมายทั่วไป"
	numLabel := "คู่เลขพลังปานกลาง"
	soundLabel := "เสียงยาว ลุ่มลึก"
	resonanceLabel := "เอกลักษณ์เฉพาะตัว"

	if h.dimensionRepo != nil {
		meaningLabel = h.dimensionRepo.GetLabelForScore(c.Request.Context(), "meaning", meaningScore)
		numLabel = h.dimensionRepo.GetLabelForScore(c.Request.Context(), "numerology", numScore)
		soundLabel = h.dimensionRepo.GetLabelForScore(c.Request.Context(), "sound", soundScore)
		resonanceLabel = h.dimensionRepo.GetLabelForScore(c.Request.Context(), "resonance", resonanceScore)
	} else {
		if meaningScore >= 90 {
			meaningLabel = "โดดเด่นระดับสูง"
		} else if meaningScore >= 80 {
			meaningLabel = "เด่นมาก มงคลลึกซึ้ง"
		} else if meaningScore >= 70 {
			meaningLabel = "ความหมายดี มงคล"
		}

		if numScore >= 90 {
			numLabel = "คู่มงคลชั้นสูง มหาบารมี"
		} else if numScore >= 80 {
			numLabel = "คู่มงคลพลังดีเยี่ยม"
		} else if numScore >= 70 {
			numLabel = "คู่มงคลเกณฑ์ดี"
		}

		if soundScore >= 90 {
			soundLabel = "จังหวะเสียงไพเราะก้องกังวาน"
		} else if soundScore >= 80 {
			soundLabel = "เสียงสง่างาม มีน้ำหนัก"
		} else if soundScore >= 70 {
			soundLabel = "ไพเราะ คล่องตัว"
		}

		if resonanceScore >= 88 {
			resonanceLabel = "สมดุลรอบด้าน เอกลักษณ์สูง"
		} else if resonanceScore >= 75 {
			resonanceLabel = "กลมกลืน ทรงเอกลักษณ์"
		}
	}

	profile4D := gin.H{
		"numerology_score": math.Round(numScore),
		"numerology_label": numLabel,
		"meaning_score":    math.Round(meaningScore),
		"meaning_label":    meaningLabel,
		"sound_score":      math.Round(soundScore),
		"sound_label":      soundLabel,
		"resonance_score":  math.Round(resonanceScore),
		"resonance_label":  resonanceLabel,
	}

	c.JSON(http.StatusOK, gin.H{
		"error":          false,
		"pairs":          pairs,
		"pair_types":     pairTypes,
		"pair_points":    pairPoints,
		"pair_is_good":   pairIsGood,
		"pair_descs":     pairDescs,
		"meaning":        combinedMeaning,
		"profile_4d":     profile4D,
		"profile_source": "exploratory_decode_v1",
		"kalakini":       kalakiniForDay(dayOfBirth, name),
	})
}

// kalakiniForDay stays in the Go domain boundary. A blank birthday deliberately
// returns no warning so users can choose a good name before providing it.
func kalakiniForDay(day, name string) gin.H {
	lettersByDay := map[string]string{
		"sunday": "ศษสหฬฮ", "monday": "ะาิีึืุูเแโใไ",
		"tuesday": "กขคฆง", "wednesday1": "จฉชซฌญ",
		"wednesday2": "บปผฝพฟภม", "thursday": "ดตถทธน",
		"friday": "ยรลว", "saturday": "ฎฏฐฑฒณ",
	}
	letters := lettersByDay[day]
	if letters == "" {
		return gin.H{"applied": false, "letters": "", "matches": ""}
	}
	set := make(map[rune]struct{}, len([]rune(letters)))
	for _, r := range letters {
		set[r] = struct{}{}
	}
	seen := make(map[rune]struct{})
	var matches []rune
	for _, r := range name {
		if _, found := set[r]; found {
			if _, duplicate := seen[r]; !duplicate {
				matches = append(matches, r)
				seen[r] = struct{}{}
			}
		}
	}
	return gin.H{"applied": true, "letters": letters, "matches": string(matches)}
}

func (h *NameHandler) GetPairDetail(c *gin.Context) {
	pairNumber := strings.TrimSpace(c.Query("number"))
	if pairNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "number is required"})
		return
	}

	allNumsMap, err := h.useCase.GetAllNumbersMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to load number dictionary"})
		return
	}

	meaning, found := allNumsMap[pairNumber]
	if !found {
		if numInt, err := strconv.Atoi(pairNumber); err == nil {
			meaning, found = allNumsMap[fmt.Sprintf("%d", numInt)]
			if !found {
				meaning, found = allNumsMap[fmt.Sprintf("%02d", numInt)]
			}
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": true, "message": "Pair detail not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":          false,
		"pair_number":    meaning.PairNumber,
		"pair_type":      meaning.PairType,
		"pair_point":     meaning.PairPoint,
		"miracle_desc":   meaning.MiracleDesc,
		"miracle_detail": meaning.MiracleDetail,
		"is_good":        meaning.IsGood(),
	})
}
