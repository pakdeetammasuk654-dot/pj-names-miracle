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
	"go-backend/usecase"
)

type entitlementStub struct{ authorized bool }

func (s entitlementStub) VerifyPurchase(context.Context, usecase.StorePurchaseRequest) (usecase.EntitlementResult, error) {
	return usecase.EntitlementResult{}, nil
}
func (s entitlementStub) Authorize(string) bool         { return s.authorized }
func (s entitlementStub) AuthorizeVerified(string) bool { return s.authorized }
func (s entitlementStub) Configured() bool              { return true }

type searchOnlyUseCase struct {
	domain.NameSearchUseCase
	result domain.NameSearchResult
	err    error
}

func (s *searchOnlyUseCase) SuggestNames(context.Context, string, int) ([]domain.NameResultItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result.Data, nil
}

type decodeErrorUseCase struct {
	domain.NameSearchUseCase
}

type randomMeaningsUseCase struct {
	domain.NameSearchUseCase
	meanings []string
	limit    int
}

func (s *randomMeaningsUseCase) GetRandomMeanings(_ context.Context, limit int) ([]string, error) {
	s.limit = limit
	return s.meanings, nil
}

func TestGetRandomMeaningsReturnsDatabaseMeanings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &randomMeaningsUseCase{meanings: []string{"ความหมายหนึ่ง", "ความหมายสอง"}}
	router := gin.New()
	router.GET("/meanings/random", NewNameHandler(stub, nil).GetRandomMeanings)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/meanings/random?limit=2", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if stub.limit != 2 {
		t.Fatalf("expected repository limit 2, got %d", stub.limit)
	}
	if !strings.Contains(recorder.Body.String(), "ความหมายหนึ่ง") {
		t.Fatalf("expected response to contain a random meaning, got %s", recorder.Body.String())
	}
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

func (s *searchOnlyUseCase) SearchTopNames(context.Context, domain.NameSearchRequest) ([]domain.NameResultItem, error) {
	return s.result.Data, s.err
}

func TestSuggestNamesRequiresEntitlementForFullNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &searchOnlyUseCase{result: domain.NameSearchResult{Data: []domain.NameResultItem{
		{THName: "เจนีตา", NameID: 42, Meaning: "ความหมายลับ", CandidateSatPairs: []domain.PairPresentation{{Value: "24", Type: "D10", IsGood: true}}, CandidateShaPairs: []domain.PairPresentation{{Value: "46", Type: "D10", IsGood: true}}},
	}}}
	for _, tc := range []struct {
		name       string
		unlocked   bool
		wantHidden bool
	}{
		{name: "free", wantHidden: true},
		{name: "paid", unlocked: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewNameHandler(stub, nil)
			h.SetEntitlementService(entitlementStub{authorized: tc.unlocked})
			router := gin.New()
			router.GET("/suggestions", h.SuggestNames)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/suggestions?query=%E0%B9%80%E0%B8%88%E0%B8%99%E0%B8%B5", nil)
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
			}
			var body struct {
				Data            []domain.NameResultItem `json:"data"`
				PreviewPairs    []suggestionPairPreview `json:"preview_pairs"`
				Count           int                     `json:"count"`
				PremiumUnlocked bool                    `json:"premium_unlocked"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Count != 1 || body.PremiumUnlocked != tc.unlocked {
				t.Fatalf("wrong entitlement envelope: %+v", body)
			}
			if tc.wantHidden {
				if len(body.Data) != 0 || len(body.PreviewPairs) != 1 || strings.Contains(recorder.Body.String(), "เจนีตา") || strings.Contains(recorder.Body.String(), "ความหมายลับ") || strings.Contains(recorder.Body.String(), "42") {
					t.Fatalf("free response leaked a name: %s", recorder.Body.String())
				}
				if body.PreviewPairs[0].MaskedName == "" || !strings.HasSuffix(body.PreviewPairs[0].MaskedName, "***") {
					t.Fatalf("expected masked name with ***, got: %q", body.PreviewPairs[0].MaskedName)
				}
			} else if len(body.Data) != 1 || body.Data[0].THName != "เจนีตา" {
				t.Fatalf("paid response did not contain name: %+v", body)
			}
		})
	}
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
	router.POST("/search", NewNameHandler(stub, nil).SearchTopNames)
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

func TestSearchTopNamesRedactsPremiumRowsOnServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &searchOnlyUseCase{result: domain.NameSearchResult{
		Mode: domain.ModeStandalone,
		Data: []domain.NameResultItem{
			{THName: "หนึ่ง", Meaning: "หนึ่ง"}, {THName: "สอง", Meaning: "สอง"},
			{THName: "สาม", Meaning: "สาม"}, {THName: "สี่ทดสอบ", Meaning: "ต้องไม่ส่ง", RankingReasons: []string{"เหตุผล"}},
		},
	}}
	h := NewNameHandler(stub, nil)
	h.SetEntitlementService(entitlementStub{authorized: false})
	router := gin.New()
	router.POST("/search", h.SearchTopNames)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBufferString(`{"mode":"standalone"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	var response struct {
		PremiumUnlocked  bool                    `json:"premium_unlocked"`
		FreeVisibleCount int                     `json:"free_visible_count"`
		Data             []domain.NameResultItem `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.PremiumUnlocked || response.FreeVisibleCount != 3 || response.Data[3].THName == "สี่ทดสอบ" || response.Data[3].Meaning != "" || len(response.Data[3].RankingReasons) != 0 {
		t.Fatalf("premium row was not safely redacted: %+v", response)
	}
}

func TestDecodePairMetadataTreatsNegativeD5AsGoodButLowPoint(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"24": {PairType: "D10", PairPoint: 80},
		"40": {PairType: "D5", PairPoint: -10},
		"47": {PairType: "R5", PairPoint: -20},
	}

	types, points, good, _ := decodePairMetadata([]string{"24", "40", "47", "06"}, numbers)
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
	router.GET("/decode", NewNameHandler(&decodeErrorUseCase{}, nil).DecodeName)
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
	router.GET("/pair-detail", NewNameHandler(stub, nil).GetPairDetail)

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
