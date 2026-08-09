package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type NameHandler struct {
	useCase domain.NameSearchUseCase
}

func NewNameHandler(useCase domain.NameSearchUseCase) *NameHandler {
	return &NameHandler{useCase: useCase}
}

func (h *NameHandler) SearchTopNames(c *gin.Context) {
	var req domain.NameSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request payload: " + err.Error(),
		})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"error":                                false,
		"mode":                                 result.Mode,
		"applied_mode":                         result.Mode,
		"count":                                len(result.Data),
		"returned_count":                       len(result.Data),
		"eligible_candidate_count":             result.EligibleCandidateCount,
		"data":                                 result.Data,
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

	c.JSON(http.StatusOK, gin.H{
		"error":                                false,
		"mode":                                 result.Mode,
		"applied_mode":                         result.Mode,
		"count":                                len(result.Data),
		"returned_count":                       len(result.Data),
		"eligible_candidate_count":             result.EligibleCandidateCount,
		"data":                                 result.Data,
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

func decodePairMetadata(pairs []string, numbers map[string]domain.NumberMeaning) ([]string, []int, []bool) {
	pairTypes := make([]string, 0, len(pairs))
	pairPoints := make([]int, 0, len(pairs))
	pairIsGood := make([]bool, 0, len(pairs))
	for _, pair := range pairs {
		meaning, ok := numbers[pair]
		if !ok {
			pairTypes = append(pairTypes, "")
			pairPoints = append(pairPoints, 0)
			pairIsGood = append(pairIsGood, false)
			continue
		}
		pairTypes = append(pairTypes, meaning.PairType)
		pairPoints = append(pairPoints, meaning.PairPoint)
		pairIsGood = append(pairIsGood, meaning.IsGood())
	}
	return pairTypes, pairPoints, pairIsGood
}

func (h *NameHandler) DecodeName(c *gin.Context) {
	name := c.Query("name")
	cleanName := strings.ReplaceAll(name, " ", "")
	if strings.TrimSpace(cleanName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "name is required"})
		return
	}

	// GetSumsByTHName retains editorial meaning but its sums are canonical. Text
	// absent from names_miracle falls back to the same dictionary calculation.
	satSum, shaSum, meaning, err := h.useCase.GetSumsByTHName(c.Request.Context(), cleanName)
	if err != nil {
		satSum, shaSum, err = h.useCase.GetSumsForText(c.Request.Context(), cleanName)
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
		meaning = ""
	}

	allNumsMap, err := h.useCase.GetAllNumbersMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to load number dictionary"})
		return
	}

	pairs := append(formatTotalValueForDecode(satSum), formatTotalValueForDecode(shaSum)...)
	pairTypes, pairPoints, pairIsGood := decodePairMetadata(pairs, allNumsMap)

	c.JSON(http.StatusOK, gin.H{
		"error":        false,
		"pairs":        pairs,
		"pair_types":   pairTypes,
		"pair_points":  pairPoints,
		"pair_is_good": pairIsGood,
		"meaning":      meaning,
	})
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
