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

// satValueMap matches sat_nums table in DB
var satValueMap = map[rune]int{
	'ก': 1, 'ด': 1, 'ถ': 1, 'ท': 1, 'ภ': 1, 'ฤ': 1, 'ฦ': 1, 'า': 1, 'ำ': 1, 'ุ': 1, '่': 1,
	'ข': 2, 'ง': 2, 'ช': 2, 'บ': 2, 'ป': 2, 'ู': 2, 'เ': 2, 'แ': 2, '้': 2,
	'ฆ': 3, 'ฑ': 3, 'ฒ': 3, 'ต': 3, '๋': 3,
	'ค': 4, 'ญ': 4, 'ธ': 4, 'ร': 4, 'ษ': 4, 'ะ': 4, 'ั': 4, 'ิ': 4, 'โ': 4,
	'ฉ': 5, 'ฌ': 5, 'ฎ': 5, 'ณ': 5, 'น': 5, 'ม': 5, 'ห': 5, 'ฬ': 5, 'ฮ': 5, 'ึ': 5,
	'จ': 6, 'ล': 6, 'ว': 6, 'อ': 6, 'ใ': 6,
	'ซ': 7, 'ศ': 7, 'ส': 7, 'ี': 7, 'ื': 7, '๊': 7,
	'ผ': 8, 'ฝ': 8, 'พ': 8, 'ฟ': 8, 'ย': 8, '็': 8,
	'ฏ': 9, 'ฐ': 9, 'ไ': 9, '์': 9,
}

// shaValueMap matches sha_nums table in DB (shadow power - different scale)
var shaValueMap = map[rune]int{
	'อ': 6, 'ะ': 6, 'า': 6, 'ิ': 6, 'ี': 6, 'ุ': 6, 'ู': 6, 'เ': 6, 'โ': 6,
	'จ': 8, 'ฉ': 8, 'ช': 8, 'ซ': 8, 'ฌ': 8, 'ญ': 8,
	'ด': 10, 'ต': 10, 'ถ': 10, 'ท': 10, 'ธ': 10, 'น': 10,
	'ย': 12, 'ร': 12, 'ล': 12, 'ว': 12,
	'ก': 15, 'ข': 15, 'ค': 15, 'ฆ': 15, 'ง': 15,
	'ฎ': 17, 'ฏ': 17, 'ฐ': 17, 'ฑ': 17, 'ฒ': 17, 'ณ': 17,
	'บ': 19, 'ป': 19, 'ผ': 19, 'ฝ': 19, 'พ': 19, 'ฟ': 19, 'ภ': 19, 'ม': 19,
	'ศ': 21, 'ษ': 21, 'ส': 21, 'ห': 21, 'ฬ': 21, 'ฮ': 21,
}

// Keep charValueMap as alias for backward compatibility
var charValueMap = satValueMap

var goodPairsMap = map[string]bool{
	"14": true, "41": true, "15": true, "51": true, "19": true, "91": true, "24": true, "42": true, "26": true, "62": true, "28": true, "82": true, "29": true, "92": true,
	"36": true, "63": true, "39": true, "93": true, "44": true, "45": true, "54": true, "46": true, "64": true, "49": true, "94": true, "55": true, "56": true, "65": true,
	"59": true, "95": true, "66": true, "69": true, "96": true, "78": true, "87": true, "79": true, "97": true, "89": true, "98": true, "99": true,
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

	var pairs []string
	var pairTypes []string
	var meaningResult string

	// Get SatSum and ShaSum from DB
	satSum, shaSum, meaning, err := h.useCase.GetSumsByTHName(c.Request.Context(), cleanName)
	if err == nil {
		allNumsMap, _ := h.useCase.GetAllNumbersMap(c.Request.Context())
		meaningResult = meaning
		
		// Process SatSum
		satPairs := formatTotalValueForDecode(satSum)
		for _, p := range satPairs {
			pairs = append(pairs, p)
			pInt, _ := strconv.Atoi(p)
			if meaningObj, ok := allNumsMap[pInt]; ok {
				pairTypes = append(pairTypes, meaningObj.PairType)
			} else {
				if goodPairsMap[p] {
					pairTypes = append(pairTypes, "D")
				} else {
					pairTypes = append(pairTypes, "R")
				}
			}
		}

		// Process ShaSum
		shaPairs := formatTotalValueForDecode(shaSum)
		for _, p := range shaPairs {
			pairs = append(pairs, p)
			pInt, _ := strconv.Atoi(p)
			if meaningObj, ok := allNumsMap[pInt]; ok {
				pairTypes = append(pairTypes, meaningObj.PairType)
			} else {
				if goodPairsMap[p] {
					pairTypes = append(pairTypes, "D")
				} else {
					pairTypes = append(pairTypes, "R")
				}
			}
		}
	} else {
		// Fallback: calculate sat_sum and sha_sum using DB lookup maps
		satSum := 0
		shaSum := 0
		for _, r := range cleanName {
			if val, ok := satValueMap[r]; ok {
				satSum += val
			}
			if val, ok := shaValueMap[r]; ok {
				shaSum += val
			}
		}
		allNumsMap, _ := h.useCase.GetAllNumbersMap(c.Request.Context())
		appendPairs := func(sum int) {
			for _, p := range formatTotalValueForDecode(sum) {
				pairs = append(pairs, p)
				pInt, _ := strconv.Atoi(p)
				if mo, ok := allNumsMap[pInt]; ok {
					pairTypes = append(pairTypes, mo.PairType)
				} else if goodPairsMap[p] {
					pairTypes = append(pairTypes, "D")
				} else {
					pairTypes = append(pairTypes, "R")
				}
			}
		}
		if satSum > 0 {
			appendPairs(satSum)
		}
		if shaSum > 0 {
			appendPairs(shaSum)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"pairs":      pairs,
		"pair_types": pairTypes,
		"meaning":    meaningResult,
	})
}

