package handler

import (
	"fmt"
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

	results, err := h.useCase.SearchTopNames(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"mode":    req.Mode,
		"count":   len(results),
		"data":    results,
		"message": "Successfully fetched top 100 auspicious names",
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
	totalStr := fmt.Sprintf("%d", total)
	if len(totalStr) < 2 {
		return []string{}
	}
	var pairs []string
	for i := 0; i < len(totalStr)-1; i++ {
		pairs = append(pairs, totalStr[i:i+2])
	}
	return pairs
}

func (h *NameHandler) DecodeName(c *gin.Context) {
	name := c.Query("name")
	cleanName := strings.ReplaceAll(name, " ", "")

	// Read pre-calculated values for known names, then calculate unknown text
	// from the same database dictionaries used by all other backend flows.
	satSum, shaSum, meaning, err := h.useCase.GetSumsByTHName(c.Request.Context(), cleanName)
	if err != nil {
		satSum, shaSum, err = h.useCase.GetSumsForText(c.Request.Context(), cleanName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to decode name"})
			return
		}
		meaning = ""
	}

	allNumsMap, err := h.useCase.GetAllNumbersMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to load number dictionary"})
		return
	}

	var pairs []string
	var pairTypes []string
	appendPairs := func(sum int) {
		for _, pair := range formatTotalValueForDecode(sum) {
			pairs = append(pairs, pair)
			pairNumber, _ := strconv.Atoi(pair)
			if meaning, ok := allNumsMap[pairNumber]; ok {
				pairTypes = append(pairTypes, meaning.PairType)
			} else {
				pairTypes = append(pairTypes, "")
			}
		}
	}
	appendPairs(satSum)
	appendPairs(shaSum)

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"pairs":      pairs,
		"pair_types": pairTypes,
		"meaning":    meaning,
	})
}
