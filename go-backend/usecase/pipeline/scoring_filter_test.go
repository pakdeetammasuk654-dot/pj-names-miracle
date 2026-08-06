package pipeline

import (
	"context"
	"testing"

	"go-backend/domain"
)

func TestNormalizeNumerologyScore(t *testing.T) {
	if got := normalizeNumerologyScore(100, 2, 100); got != 50 {
		t.Fatalf("expected 50, got %.2f", got)
	}
	if got := normalizeNumerologyScore(300, 2, 100); got != 100 {
		t.Fatalf("expected score to clamp at 100, got %.2f", got)
	}
}

func TestScoringFilterSemanticWeightCanOutrankPerfectBonus(t *testing.T) {
	semanticWeight := 1.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			WishQuery:      "ผู้นำ",
			SemanticWeight: &semanticWeight,
			Limit:          10,
		},
		GoodNumbers: map[int]domain.NumberMeaning{
			24: {PairType: "D10", PairPoint: 100},
			42: {PairType: "D10", PairPoint: 100},
			12: {PairType: "R5", PairPoint: 10},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "คู่เลขสมบูรณ์", SatSum: 24, ShaSum: 42, VectorDistance: 0.9},
			{NameID: 2, THName: "ความหมายตรง", SatSum: 24, ShaSum: 12, VectorDistance: 0.0},
		},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.FinalResults[0].NameID != 2 {
		t.Fatalf("semantic 100%% should rank the best semantic match first, got %q", payload.FinalResults[0].THName)
	}
	if payload.FinalResults[1].PerfectBonus != perfectNumerologyBonus {
		t.Fatalf("expected transparent perfect bonus %.0f, got %.0f", perfectNumerologyBonus, payload.FinalResults[1].PerfectBonus)
	}
}
