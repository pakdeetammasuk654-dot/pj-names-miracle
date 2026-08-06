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

func TestScoringFilterExcludesBadPairEvenWhenSemanticMatchIsPerfect(t *testing.T) {
	semanticWeight := 1.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:           domain.ModeCombined,
			WishQuery:      "ผู้นำ",
			SemanticWeight: &semanticWeight,
			Limit:          10,
		},
		SurnameSatSum: 23,
		GoodNumbers: map[int]domain.NumberMeaning{
			24: {PairType: "D10", PairPoint: 100},
			42: {PairType: "D10", PairPoint: 100},
			47: {PairType: "R5", PairPoint: -50},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "คู่เลขมงคล", SatSum: 1, ShaSum: 42, VectorDistance: 0.9},
			{NameID: 2, THName: "ความหมายตรงแต่มีคู่ร้าย", SatSum: 24, ShaSum: 42, VectorDistance: 0.0},
		},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.FinalResults) != 1 {
		t.Fatalf("expected only the all-good candidate, got %d results", len(payload.FinalResults))
	}
	if payload.FinalResults[0].NameID != 1 {
		t.Fatalf("expected bad-pair candidate to be excluded, got %q", payload.FinalResults[0].THName)
	}
	if payload.FinalResults[0].PerfectBonus != perfectNumerologyBonus {
		t.Fatalf("expected transparent perfect bonus %.0f, got %.0f", perfectNumerologyBonus, payload.FinalResults[0].PerfectBonus)
	}
}

func TestAllPairsAreGoodRejectsOneBadPairInMultiDigitTotal(t *testing.T) {
	numbers := map[int]domain.NumberMeaning{
		14: {PairType: "D8"},
		47: {PairType: "R5"},
	}

	if allPairsAreGood(formatTotalValue(147), numbers) {
		t.Fatal("expected total 147 to be rejected because pair 47 is bad")
	}
}

func TestEligibleBaseSumsUsesPostSurnameTotalsAcrossFullRange(t *testing.T) {
	numbers := map[int]domain.NumberMeaning{
		14: {PairType: "D8"},
		24: {PairType: "D10"},
		42: {PairType: "D10"},
		47: {PairType: "R5"},
	}

	eligible := eligibleBaseSums(23, 120, numbers)
	got := make(map[int]bool, len(eligible))
	for _, sum := range eligible {
		got[sum] = true
	}

	if !got[1] { // 1 + 23 = 24 (D10)
		t.Fatal("expected base sum 1 to be eligible after surname sum 23")
	}
	if !got[19] { // 19 + 23 = 42 (D10)
		t.Fatal("expected base sum 19 to be eligible after surname sum 23")
	}
	if got[24] { // 24 + 23 = 47 (R5)
		t.Fatal("expected base sum 24 to be excluded after surname sum 23")
	}
	if !got[119] { // 119 + 23 = 142 -> 14, 42 (both D)
		t.Fatal("expected multi-digit total 142 to be eligible")
	}
}
