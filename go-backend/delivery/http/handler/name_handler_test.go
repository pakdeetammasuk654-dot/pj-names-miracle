package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type searchOnlyUseCase struct {
	domain.NameSearchUseCase
	result domain.NameSearchResult
	err    error
}

type decodeErrorUseCase struct {
	domain.NameSearchUseCase
}

func (s *decodeErrorUseCase) GetSumsByTHName(context.Context, string) (int, int, string, error) {
	return 0, 0, "", domain.ErrUnmappedCharacters
}

func (s *decodeErrorUseCase) GetSumsForText(context.Context, string) (int, int, error) {
	return 0, 0, errors.Join(domain.ErrUnmappedCharacters, errors.New("unsupported rune"))
}

func (s *searchOnlyUseCase) SearchTopNamesWithToken(context.Context, domain.NameSearchRequest) (domain.NameSearchResult, error) {
	return s.result, s.err
}

func TestSearchTopNamesKeepsLegacyEnvelopeAndAddsRankingMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &searchOnlyUseCase{result: domain.NameSearchResult{
		CandidateToken:                    "opaque-token",
		Mode:                              domain.ModeStandalone,
		Data:                              []domain.NameResultItem{{NameID: 1, THName: "ทดสอบ"}},
		EligibleCandidateCount:            1234,
		RequestedWeights:                  domain.RankingWeights{Numerology: 0.15, Semantic: 0.15, Sound: 0.70},
		AppliedWeights:                    domain.RankingWeights{Numerology: 0.15, Semantic: 0.15, Sound: 0.70},
		SemanticAvailable:                 true,
		SemanticSource:                    "embedding",
		SemanticQueryModel:                "text-embedding-3-small",
		SemanticModel:                     "text-embedding-3-small",
		SemanticVectorHash:                "abc123",
		CandidateVectorProvenanceVerified: true,
		CandidateVectorProvenanceSource:   "approved_snapshot",
		Policy: domain.NameSearchPolicy{
			PairScope:                 "first_name_all_d_pairs",
			KalakiniScope:             "not_applied",
			SoundScope:                "first_name",
			NumerologyRuleVersion:     domain.NumerologyRuleVersion,
			SoundFormulaVersion:       domain.SoundFormulaVersion,
			SoundCalibrationSource:    "live_repository_distribution_provisional",
			SoundCalibrationReference: "v1",
			SoundCalibrationApproved:  false,
			CorpusFilterVersion:       domain.CorpusFilterVersion,
		},
	}}
	router := gin.New()
	router.POST("/search", NewNameHandler(stub).SearchTopNames)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBufferString(`{"mode":"standalone"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, field := range []string{
		"mode", "count", "data", "message",
		"applied_mode", "returned_count", "eligible_candidate_count",
		"candidate_token", "requested_weights", "applied_weights",
		"semantic_available", "semantic_source", "semantic_model", "semantic_query_model", "semantic_vector_hash",
		"candidate_vector_provenance_verified", "candidate_vector_provenance_source", "policy",
	} {
		if _, ok := response[field]; !ok {
			t.Fatalf("response is missing backward-compatible/metadata field %q: %s", field, recorder.Body.String())
		}
	}
	var message string
	if err := json.Unmarshal(response["message"], &message); err != nil || message != "Successfully fetched 1 auspicious names" {
		t.Fatalf("expected dynamic count in success message, got %q (err=%v)", message, err)
	}
}

func TestDecodePairMetadataTreatsNegativeD5AsGoodButLowPoint(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"24": {PairType: "D10", PairPoint: 80},
		"40": {PairType: "D5", PairPoint: -10},
		"47": {PairType: "R5", PairPoint: -20},
	}

	types, points, good := decodePairMetadata([]string{"24", "40", "47", "06"}, numbers)
	if !reflect.DeepEqual(types, []string{"D10", "D5", "R5", ""}) {
		t.Fatalf("unexpected pair types: %v", types)
	}
	if !reflect.DeepEqual(points, []int{80, -10, -20, 0}) {
		t.Fatalf("unexpected pair points: %v", points)
	}
	if !reflect.DeepEqual(good, []bool{true, true, false, false}) {
		t.Fatalf("expected D5 negative point to remain good while R and unknown fail: %v", good)
	}
}

func TestFormatTotalValueForDecodeUsesLeadingZeroForSingleDigitTotal(t *testing.T) {
	if got := formatTotalValueForDecode(6); !reflect.DeepEqual(got, []string{"06"}) {
		t.Fatalf("expected single-digit total to decode through pair 06, got %v", got)
	}
}

func TestDecodeNameReturnsBadRequestForUnmappedCharacters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/decode", NewNameHandler(&decodeErrorUseCase{}).DecodeName)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/decode?name=%F0%9F%99%82", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unmapped characters, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

type stubPairUseCase struct {
	domain.NameSearchUseCase
	allNumsMap map[string]domain.NumberMeaning
}

func (s *stubPairUseCase) GetAllNumbersMap(context.Context) (map[string]domain.NumberMeaning, error) {
	return s.allNumsMap, nil
}

func TestGetPairDetailReturnsMeaning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubPairUseCase{
		allNumsMap: map[string]domain.NumberMeaning{
			"64": {PairNumber: "64", PairType: "D10", PairPoint: 80, MiracleDesc: "ความสำเร็จมหาโชค", MiracleDetail: "รายละเอียดแบบเต็มของ 64"},
		},
	}
	router := gin.New()
	router.GET("/pair-detail", NewNameHandler(stub).GetPairDetail)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/pair-detail?number=64", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if !strings.Contains(recorder.Body.String(), "ความสำเร็จมหาโชค") {
		t.Fatalf("expected body to contain MiracleDesc, got %s", recorder.Body.String())
	}
}
