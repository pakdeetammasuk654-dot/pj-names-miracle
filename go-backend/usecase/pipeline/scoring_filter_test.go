package pipeline

import (
	"context"
	"strings"
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

func TestRankingPrecisionRoundsAtTwelveDecimalPlaces(t *testing.T) {
	base := 50.0
	if got := roundRankingScore(base + 0.4e-12); got != base {
		t.Fatalf("sub-precision noise must round to the same score, got %.15f", got)
	}
	if got := roundRankingScore(base + 0.6e-12); got != base+1e-12 {
		t.Fatalf("value beyond half a precision unit must remain distinguishable, got %.15f", got)
	}
}

func TestSubPrecisionSemanticNoiseFallsThroughToStableNameTieBreaker(t *testing.T) {
	semanticWeight := 1.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:           domain.ModeStandalone,
			WishQuery:      "ความหมาย",
			SemanticWeight: &semanticWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		SemanticCorpus: domain.SemanticCorpusProvenance{Approved: true, Source: "approved_snapshot"},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "ข", SatSum: 24, ShaSum: 24, VectorDistance: 0.1, SemanticAvailable: true, SemanticSource: "embedding"},
			{NameID: 2, THName: "ก", SatSum: 24, ShaSum: 24, VectorDistance: 0.1 + 1e-15, SemanticAvailable: true, SemanticSource: "embedding"},
		},
	}

	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score candidates: %v", err)
	}
	if payload.FinalResults[0].THName != "ก" {
		t.Fatalf("sub-precision noise bypassed stable Thai-name tie-break: first=%q", payload.FinalResults[0].THName)
	}
}

func TestScoringFilterKeepsKnownBadPairAndSignalsCaution(t *testing.T) {
	numerologyWeight := 0.50
	semanticWeight := 0.50
	soundWeight := 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:             domain.ModeCombined,
			WishQuery:        "ผู้นำ",
			NumerologyWeight: &numerologyWeight,
			SemanticWeight:   &semanticWeight,
			SoundWeight:      &soundWeight,
			Limit:            10,
		},
		SurnameSatSum: 18,
		SurnameShaSum: 18,
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"29": {PairType: "D8", PairPoint: 70},
			"42": {PairType: "D10", PairPoint: 100},
			"47": {PairType: "R5", PairPoint: -50},
		},
		SemanticCorpus: domain.SemanticCorpusProvenance{Approved: true, Source: "approved_snapshot"},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "คู่เลขมงคล", SatSum: 24, ShaSum: 24, VectorDistance: 0.9, SemanticAvailable: true, SemanticSource: "embedding", SemanticQueryModel: "text-embedding-3-small", SemanticVectorHash: "abc123"},
			{NameID: 2, THName: "ความหมายตรงแต่มีคู่ร้าย", SatSum: 29, ShaSum: 24, VectorDistance: 0.0, SemanticAvailable: true, SemanticSource: "embedding", SemanticQueryModel: "text-embedding-3-small", SemanticVectorHash: "abc123"},
		},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.FinalResults) != 2 {
		t.Fatalf("expected every governed candidate, got %d results", len(payload.FinalResults))
	}
	bad := payload.FinalResults[0]
	if bad.NameID != 2 || !bad.HasBadPair || bad.PairRiskLevel != "caution_r5" {
		t.Fatalf("known bad pair must remain scoreable with a caution flag, got %+v", bad)
	}
	if payload.SemanticQueryModel != "text-embedding-3-small" || payload.SemanticVectorHash != "abc123" ||
		payload.FinalResults[0].SemanticVectorHash != "abc123" {
		t.Fatalf("semantic proof provenance was not preserved: model=%q hash=%q item=%q", payload.SemanticQueryModel, payload.SemanticVectorHash, payload.FinalResults[0].SemanticVectorHash)
	}
}

func TestScoringFilterAllowsOneR5WhenNumerologyWeightIsLow(t *testing.T) {
	numerologyWeight := 0.25
	semanticWeight := 0.75
	soundWeight := 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode: domain.ModeStandalone, WishQuery: "ผู้นำ", Limit: 10,
			NumerologyWeight: &numerologyWeight, SemanticWeight: &semanticWeight, SoundWeight: &soundWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"47": {PairType: "R5", PairPoint: -20},
		},
		SemanticCorpus: domain.SemanticCorpusProvenance{Approved: true, Source: "approved_snapshot"},
		Candidates: []domain.NameResultItem{{
			NameID: 1, THName: "ความหมายตรง", SatSum: 47, ShaSum: 24,
			VectorDistance: 0, SemanticAvailable: true, SemanticSource: "embedding",
		}},
	}
	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score candidates: %v", err)
	}
	if len(payload.FinalResults) != 1 || !payload.FinalResults[0].HasBadPair {
		t.Fatalf("expected one visible R5 candidate, got %+v", payload.FinalResults)
	}
	if payload.FinalResults[0].SatPairType != "R5" || payload.FinalResults[0].PairRiskLevel != "caution_r5" {
		t.Fatalf("red-pair metadata missing: %+v", payload.FinalResults[0])
	}
}

func TestScoringFilterKeepsKnownRiskGradesAtEveryWeight(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{"24": {PairType: "D10"}, "47": {PairType: "R7", PairPoint: -40}}
	for _, weight := range []float64{0, 0.20, 1} {
		decision := evaluateAdaptivePairPolicy([]string{"24", "47"}, numbers, weight)
		if !decision.Allowed || decision.Level != "caution_r7" {
			t.Fatalf("R7 must remain a scoreable caution at %.0f%%: %+v", weight*100, decision)
		}
	}
}

func TestZeroNumerologyAllowsGovernedR10AndMultipleBadPairs(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"13": {PairType: "R5"}, "31": {PairType: "R5"}, "67": {PairType: "R10"},
	}
	if !evaluateAdaptivePairPolicy([]string{"67"}, numbers, 0).Allowed {
		t.Fatal("R10 must remain visible when numerology is not applied")
	}
	decision := evaluateAdaptivePairPolicy([]string{"13", "31"}, numbers, 1)
	if !decision.Allowed || decision.BadCount != 2 {
		t.Fatalf("all governed pairs must remain scoreable at every weight, got %+v", decision)
	}
}

func TestAdaptivePairScopeMatchesSliderThresholds(t *testing.T) {
	weight := func(value float64) *float64 { return &value }
	tests := []struct {
		numerology float64
		want       string
	}{
		{numerology: 0.35, want: "all_known_pairs_scored"},
		{numerology: 0.20, want: "all_known_pairs_scored"},
		{numerology: 0, want: "all_known_pairs_scored"},
	}
	for _, tt := range tests {
		request := domain.NameSearchRequest{
			NumerologyWeight: weight(tt.numerology),
			SemanticWeight:   weight(1 - tt.numerology),
			SoundWeight:      weight(0),
		}
		if got := AdaptivePairScope(request); got != tt.want {
			t.Fatalf("numerology %.0f%%: want %q, got %q", tt.numerology*100, tt.want, got)
		}
	}
}

func TestZeroNumerologyTopResultsUsesNaturalScoreOrderWithoutPairQuota(t *testing.T) {
	items := []domain.NameResultItem{
		{NameID: 1, FinalScore: 100, HasBadPair: true},
		{NameID: 2, FinalScore: 99, HasBadPair: true},
		{NameID: 3, FinalScore: 98},
	}
	selected := selectAdaptiveTopResults(items, 2, 0)
	if len(selected) != 2 || selected[0].NameID != 1 || selected[1].NameID != 2 {
		t.Fatalf("0%% numerology must preserve natural non-numerology order, got %+v", selected)
	}
}

func TestTopResultsNeverReservesCautionRows(t *testing.T) {
	items := make([]domain.NameResultItem, 0, 120)
	for i := 0; i < 115; i++ {
		items = append(items, domain.NameResultItem{NameID: i + 1, FinalScore: float64(200 - i)})
	}
	for i := 0; i < 5; i++ {
		items = append(items, domain.NameResultItem{NameID: 200 + i, FinalScore: float64(80 - i), HasBadPair: true})
	}
	selected := selectAdaptiveTopResults(items, 100, 0.25)
	for i, item := range selected {
		if i > 0 && selected[i-1].FinalScore < item.FinalScore {
			t.Fatalf("selected result order changed at %d: %.2f before %.2f", i, selected[i-1].FinalScore, item.FinalScore)
		}
	}
	if selected[len(selected)-1].NameID != 100 {
		t.Fatalf("top 100 must be the highest score rows without a caution quota, got last=%+v", selected[len(selected)-1])
	}
}

func TestNumerologyOnlyScoreDoesNotReceivePerfectPairBonus(t *testing.T) {
	semanticWeight := 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:           domain.ModeStandalone,
			WishQuery:      "หอมหวาน",
			SemanticWeight: &semanticWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"45": {PairType: "D10", PairPoint: 85},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "คร้อย", SatSum: 24, ShaSum: 45, VectorDistance: 0},
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
		t.Fatalf("expected one result, got %d", len(payload.FinalResults))
	}

	const want = 165.0 / 170.0 * 100.0
	got := payload.FinalResults[0].FinalScore
	if diff := got - want; diff < -0.0001 || diff > 0.0001 {
		t.Fatalf("expected raw numerology score %.4f without bonus, got %.4f", want, got)
	}
	if got >= 100 {
		t.Fatalf("expected 24/45 to remain below 100, got %.2f", got)
	}
}

func TestThreeDimensionWeightsAreNormalized(t *testing.T) {
	numerologyWeight := 0.45
	semanticWeight := 0.35
	soundWeight := 0.20
	num, sem, snd, sim := resolveRankingWeights(domain.NameSearchRequest{
		NumerologyWeight: &numerologyWeight,
		SemanticWeight:   &semanticWeight,
		SoundWeight:      &soundWeight,
	})

	if num != 0.45 || sem != 0.35 || snd != 0.20 || sim != 0 {
		t.Errorf("default weights mismatch: got %.2f/%.2f/%.2f/%.2f", num, sem, snd, sim)
	}
}

func TestFourDimensionWeightsAndSimilarityRanking(t *testing.T) {
	zero := 0.0
	one := 1.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode: domain.ModeStandalone, PrototypeName: "ณเดชน์", IsNameIntent: true,
			NumerologyWeight: &zero, SemanticWeight: &zero, SoundWeight: &zero, SimilarityWeight: &one,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 3, THName: "ณเดชน์", SatSum: 24, ShaSum: 24, SimilarityDistance: 0},
			{NameID: 1, THName: "ชื่อคล้ายน้อย", SatSum: 24, ShaSum: 24, SimilarityDistance: 0.80},
			{NameID: 2, THName: "ชื่อคล้ายมาก", SatSum: 24, ShaSum: 24, SimilarityDistance: 0.10},
		},
	}

	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score candidates: %v", err)
	}
	if len(payload.FinalResults) != 3 || payload.FinalResults[0].NameID != 3 {
		t.Fatalf("expected exact pg_trgm match to rank first, got %+v", payload.FinalResults)
	}
	if payload.FinalResults[0].SimilarityScore != 100 || payload.AppliedWeights.Similarity != 1 {
		t.Fatalf("unexpected similarity score/weight: item=%+v weights=%+v", payload.FinalResults[0], payload.AppliedWeights)
	}
}

func TestSurnameCandidateUsesSurnameSoundAndKalakiniScope(t *testing.T) {
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:         domain.ModeCombined,
			SearchTarget: domain.TargetSurname,
			DayOfBirth:   "sunday",
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		Candidates: []domain.NameResultItem{{
			NameID: 1, THName: "นามสกุลใหม่", SatSum: 24, ShaSum: 24,
		}},
	}

	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score surname candidate: %v", err)
	}
	if len(payload.FinalResults) != 1 {
		t.Fatalf("expected one surname result, got %+v", payload.FinalResults)
	}
	item := payload.FinalResults[0]
	if item.SoundScope != "surname" {
		t.Fatalf("expected surname sound scope, got %q", item.SoundScope)
	}
	if !strings.Contains(strings.Join(item.RankingReasons, " "), "นามสกุลผ่านการกรองกาลกิณี") {
		t.Fatalf("expected candidate-role Kalakini reason, got %+v", item.RankingReasons)
	}
}

func TestCalculateSoundScoreUsesTransparentComponentWeights(t *testing.T) {
	got := calculateSoundScore(80, 90, 70, "กิตติ") // Length 5 -> 0 penalty
	const want = 81.5                               // (80*45%) + (90*35%) + (70*20%)
	if got != want {
		t.Fatalf("expected sound score %.1f, got %.1f", want, got)
	}
}

func TestCalculateSoundScoreWithLengthPenalty(t *testing.T) {
	// Raw score = 81.5
	// Length = 3 -> penalty = (5-3)*2 = 4. Final = 77.5
	got1 := calculateSoundScore(80, 90, 70, "กิต")
	if got1 != 77.5 {
		t.Fatalf("expected 77.5, got %.1f", got1)
	}

	// Length = 11 -> penalty = (11-9)*2 = 4. Final = 77.5
	got2 := calculateSoundScore(80, 90, 70, "กิตติพงศ์ธร")
	if got2 != 77.5 {
		t.Fatalf("expected 77.5, got %.1f", got2)
	}
}

func TestEqualPrimaryScoresPreferBetterSoundQuality(t *testing.T) {
	numerologyWeight := 1.0
	semanticWeight := 0.0
	soundWeight := 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:             domain.ModeStandalone,
			WishQuery:        "หอมหวาน",
			NumerologyWeight: &numerologyWeight,
			SemanticWeight:   &semanticWeight,
			SoundWeight:      &soundWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"45": {PairType: "D10", PairPoint: 85},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "เยแจ่ม", SatSum: 24, ShaSum: 45, EuphonyScore: 80, PronunciationEase: 70, RhythmScore: 75},
			{NameID: 2, THName: "คร้อย", SatSum: 24, ShaSum: 45, EuphonyScore: 85, PronunciationEase: 90, RhythmScore: 90},
		},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.FinalResults) != 2 {
		t.Fatalf("expected two results, got %d", len(payload.FinalResults))
	}
	if payload.FinalResults[0].THName != "คร้อย" {
		t.Fatalf("expected better-sounding equal-score name first, got %q", payload.FinalResults[0].THName)
	}
	if payload.FinalResults[0].FinalScore != payload.FinalResults[1].FinalScore {
		t.Fatal("sound tie-breaker must not change a 100% numerology score")
	}
}

func TestAllPairsAreGoodRejectsOneBadPairInMultiDigitTotal(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"14": {PairType: "D8", PairPoint: 50},
		"47": {PairType: "R5", PairPoint: -20},
	}

	if allPairsAreGood(formatTotalValue(147), numbers) {
		t.Fatal("expected total 147 to be rejected because pair 47 is bad")
	}
}

func TestEligibleBaseSumsUsesPostSurnameTotalsAcrossFullRange(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"01": {PairType: "D5", PairPoint: 10},
		"11": {PairType: "D8", PairPoint: 40},
		"14": {PairType: "D8", PairPoint: 50},
		"19": {PairType: "D8", PairPoint: 60},
		"24": {PairType: "D10", PairPoint: 80},
		"42": {PairType: "D10", PairPoint: 100},
		"47": {PairType: "R5", PairPoint: -20},
	}

	eligible := eligibleBaseSums(23, 120, numbers, true)
	got := make(map[int]bool, len(eligible))
	for _, sum := range eligible {
		got[sum] = true
	}

	if !got[1] { // first-name 1 (D5), then 1 + 23 = 24 (D10)
		t.Fatal("expected base sum 1 to be eligible after surname sum 23")
	}
	if !got[19] { // 19 + 23 = 42 (D10)
		t.Fatal("expected base sum 19 to be eligible after surname sum 23")
	}
	if !got[24] { // Cached superset keeps every governed grade for 0% reranking.
		t.Fatal("expected base sum 24 to remain in complete candidate pool")
	}
	if !got[119] { // 119 + 23 = 142 -> 14, 42 (both D)
		t.Fatal("expected multi-digit total 142 to be eligible")
	}
}

func TestSingleDigitTotalLooksUpLeadingZeroPair(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"6":  {PairType: "D10", PairPoint: 100},
		"06": {PairType: "R5", PairPoint: -10},
	}

	if allPairsAreGood(formatTotalValue(6), numbers) {
		t.Fatal("expected total 6 to lookup leading-zero pair 06, not distinct authored key 6")
	}
}

func TestDTypeWithNegativePointRemainsEligibleButScoresLow(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"40": {PairType: "D5", PairPoint: -10},
	}

	if !allPairsAreGood([]string{"40"}, numbers) {
		t.Fatal("expected D5 pair with a negative point to remain eligible")
	}
	if got := normalizeNumerologyScore(-10, 1, 100); got != 0 {
		t.Fatalf("expected negative D5 point to lower normalized score to zero, got %.2f", got)
	}
}

func TestEligibleStandaloneSumsIncludeValuesAboveNinetyNine(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"11": {PairType: "D8", PairPoint: 40},
		"19": {PairType: "D8", PairPoint: 60},
	}

	eligible := eligibleBaseSums(0, 120, numbers, false)
	for _, sum := range eligible {
		if sum == 119 {
			return
		}
	}
	t.Fatal("expected standalone sum 119 (pairs 11 and 19) in the complete eligible range")
}

func TestCombinedPoolIncludesGovernedBadCandidateForZeroNumerology(t *testing.T) {
	numbers := map[string]domain.NumberMeaning{
		"01": {PairType: "R5", PairPoint: -10},
		"24": {PairType: "D10", PairPoint: 80},
	}

	eligible := eligibleBaseSums(23, 1, numbers, true)
	if len(eligible) != 1 || eligible[0] != 1 {
		t.Fatalf("expected governed bad candidate in complete rerank pool, got %v", eligible)
	}
}

func TestCombinedBadCandidateIsScoreableAtEveryNumerologyWeight(t *testing.T) {
	zero, one := 0.0, 1.0
	newPayload := func(numerology, sound *float64) *domain.NameSearchPayload {
		return &domain.NameSearchPayload{
			Request: domain.NameSearchRequest{
				Mode: domain.ModeCombined, NumerologyWeight: numerology,
				SemanticWeight: &zero, SoundWeight: sound, SimilarityWeight: &zero,
			},
			SurnameSatSum: 23,
			SurnameShaSum: 23,
			GoodNumbers: map[string]domain.NumberMeaning{
				"01": {PairType: "R5", PairPoint: -10},
				"24": {PairType: "D10", PairPoint: 80},
			},
			Candidates: []domain.NameResultItem{{NameID: 1, THName: "ทดสอบ", SatSum: 1, ShaSum: 1}},
		}
	}

	zeroPayload := newPayload(&zero, &one)
	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: zeroPayload}); err != nil {
		t.Fatalf("score 0%% numerology: %v", err)
	}
	if len(zeroPayload.FinalResults) != 1 || zeroPayload.FinalResults[0].BadPairCount != 0 {
		// The displayed combined pair is 24 (D10); base pair 01 is intentionally
		// not used as a hidden rejection when numerology is disabled.
		t.Fatalf("expected candidate at 0%% numerology, got %+v", zeroPayload.FinalResults)
	}

	positivePayload := newPayload(&one, &zero)
	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: positivePayload}); err != nil {
		t.Fatalf("score positive numerology: %v", err)
	}
	if len(positivePayload.FinalResults) != 1 {
		t.Fatalf("expected governed candidate to remain scoreable above 0%%, got %+v", positivePayload.FinalResults)
	}
}

func TestCombinedAdaptiveWarningAlwaysMapsToDisplayedBadPair(t *testing.T) {
	numerologyWeight, semanticWeight, soundWeight := 0.25, 0.75, 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode: domain.ModeCombined, Limit: 10,
			NumerologyWeight: &numerologyWeight, SemanticWeight: &semanticWeight, SoundWeight: &soundWeight,
		},
		SurnameSatSum: 23,
		SurnameShaSum: 23,
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"47": {PairType: "R5", PairPoint: -20},
		},
		Candidates: []domain.NameResultItem{{NameID: 1, THName: "ทดสอบ", SatSum: 24, ShaSum: 24}},
	}
	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score combined candidate: %v", err)
	}
	if len(payload.FinalResults) != 1 || payload.FinalResults[0].BadPairCount != 2 {
		t.Fatalf("two displayed red pairs must remain scoreable and flagged, got %+v", payload.FinalResults)
	}

	// Only one post-surname dimension becomes red; the warning must therefore
	// correspond exactly to the red pair included in the response display.
	payload.Candidates = []domain.NameResultItem{{NameID: 2, THName: "แสดงแดง", SatSum: 24, ShaSum: 1}}
	payload.GoodNumbers["01"] = domain.NumberMeaning{PairType: "D8", PairPoint: 50}
	if err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{Ctx: context.Background(), Data: payload}); err != nil {
		t.Fatalf("score visible caution candidate: %v", err)
	}
	if len(payload.FinalResults) != 1 || !payload.FinalResults[0].HasBadPair || payload.FinalResults[0].SatPairType != "R5" {
		t.Fatalf("expected warning backed by displayed R5 pair, got %+v", payload.FinalResults)
	}
}

func TestSemanticUnavailableRedistributesWeightWithoutFakePerfectScore(t *testing.T) {
	numerologyWeight := 0.45
	semanticWeight := 0.35
	soundWeight := 0.20
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:             domain.ModeStandalone,
			WishQuery:        "ผู้นำ",
			NumerologyWeight: &numerologyWeight,
			SemanticWeight:   &semanticWeight,
			SoundWeight:      &soundWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"45": {PairType: "D10", PairPoint: 100},
		},
		Candidates: []domain.NameResultItem{{
			NameID:            1,
			THName:            "ชื่อทดสอบ",
			SatSum:            24,
			ShaSum:            45,
			VectorDistance:    0,
			SemanticAvailable: false,
			SemanticSource:    "unavailable",
			EuphonyScore:      80,
			PronunciationEase: 80,
			RhythmScore:       80,
		}},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.FinalResults) != 1 {
		t.Fatalf("expected one result, got %d", len(payload.FinalResults))
	}
	item := payload.FinalResults[0]
	if item.SemanticAvailable || item.SemanticScore != 0 {
		t.Fatalf("semantic fallback must stay unavailable with score zero, got available=%v score=%.2f", item.SemanticAvailable, item.SemanticScore)
	}
	if payload.AppliedWeights.Semantic != 0 {
		t.Fatalf("expected semantic weight to be redistributed, got %.4f", payload.AppliedWeights.Semantic)
	}
	if item.FinalScore >= 100 {
		t.Fatalf("fallback must not manufacture a perfect final score, got %.2f", item.FinalScore)
	}
	if payload.SemanticSource != "unavailable" {
		t.Fatalf("expected honest semantic source, got %q", payload.SemanticSource)
	}
}

func TestScoringFailsClosedWhenCandidateVectorsLackApprovedCorpusProof(t *testing.T) {
	semanticWeight := 1.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:           domain.ModeStandalone,
			WishQuery:      "ความหมาย",
			SemanticWeight: &semanticWeight,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		SemanticCorpus: domain.SemanticCorpusProvenance{
			Approved: false,
			Source:   "unavailable_unapproved_vector_corpus",
		},
		Candidates: []domain.NameResultItem{{
			NameID:             1,
			THName:             "ชื่อทดสอบ",
			SatSum:             24,
			ShaSum:             24,
			VectorDistance:     0,
			SemanticAvailable:  true,
			SemanticSource:     "embedding",
			SemanticQueryModel: domain.SemanticEmbeddingModel,
			SemanticVectorHash: strings.Repeat("a", 64),
		}},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.SemanticAvailable || payload.FinalResults[0].SemanticAvailable ||
		payload.FinalResults[0].SemanticScore != 0 {
		t.Fatal("candidate-reported embedding must not bypass an unapproved corpus contract")
	}
	if payload.CandidateVectorProvenanceVerified ||
		payload.CandidateVectorProvenanceSource != "unavailable_unapproved_vector_corpus" {
		t.Fatalf("unexpected fail-closed provenance: verified=%v source=%q", payload.CandidateVectorProvenanceVerified, payload.CandidateVectorProvenanceSource)
	}
	if payload.SemanticSource != "unavailable_unapproved_vector_corpus" ||
		payload.SemanticQueryModel != "" || payload.SemanticVectorHash != "" {
		t.Fatalf("unapproved candidate claims leaked into response proof: source=%q model=%q hash=%q", payload.SemanticSource, payload.SemanticQueryModel, payload.SemanticVectorHash)
	}
}

func TestCalibrateSoundScoreMapsVersionPercentileToReferenceDistribution(t *testing.T) {
	profile := domain.SoundCalibrationProfile{
		ReferenceVersion: "v1",
		ReferenceMean:    80,
		ReferenceStdDev:  10,
		Source:           "live_repository_distribution_provisional",
		Versions: map[string]domain.SoundCalibrationStats{
			"v1": {Count: 1000, Mean: 80, StdDev: 10},
			"v2": {Count: 200, Mean: 70, StdDev: 5},
		},
	}

	got, source := calibrateSoundScore(75, "v2", profile)
	if got != 90 {
		t.Fatalf("expected v2 score one standard deviation high to calibrate to 90, got %.2f", got)
	}
	if source != "live_zscore_to_reference:v1" {
		t.Fatalf("unexpected calibration source %q", source)
	}
}

func TestCalibrateSoundScoreLabelsApprovedSnapshot(t *testing.T) {
	profile := domain.SoundCalibrationProfile{
		ReferenceVersion: "v1",
		ReferenceMean:    80,
		ReferenceStdDev:  10,
		Source:           "approved_snapshot",
		Approved:         true,
		Versions: map[string]domain.SoundCalibrationStats{
			"v1": {Count: 1000, Mean: 80, StdDev: 10},
		},
	}

	_, source := calibrateSoundScore(90, "v1", profile)
	if source != "approved_snapshot_zscore_to_reference:v1" {
		t.Fatalf("unexpected approved calibration source %q", source)
	}
}

func TestEqualCalibratedSoundDoesNotReintroduceRawVersionBias(t *testing.T) {
	one := 1.0
	zero := 0.0
	payload := &domain.NameSearchPayload{
		Request: domain.NameSearchRequest{
			Mode:             domain.ModeStandalone,
			NumerologyWeight: &one,
			SemanticWeight:   &zero,
			SoundWeight:      &zero,
		},
		GoodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		SoundCalibration: domain.SoundCalibrationProfile{
			ReferenceVersion: "v1",
			ReferenceMean:    80,
			ReferenceStdDev:  10,
			Source:           "live_repository_distribution_provisional",
			Versions: map[string]domain.SoundCalibrationStats{
				"v1": {Count: 100, Mean: 80, StdDev: 10},
				"v2": {Count: 100, Mean: 70, StdDev: 5},
			},
		},
		Candidates: []domain.NameResultItem{
			{NameID: 1, THName: "กขคงจ", SatSum: 24, ShaSum: 24, PhoneticVersion: "v2", EuphonyScore: 75, PronunciationEase: 75, RhythmScore: 75},
			{NameID: 2, THName: "กขคงฉ", SatSum: 24, ShaSum: 24, PhoneticVersion: "v1", EuphonyScore: 90, PronunciationEase: 90, RhythmScore: 90},
		},
	}

	err := NewScoringFilter().Execute(&domain.PipelineContext[*domain.NameSearchPayload]{
		Ctx:  context.Background(),
		Data: payload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload.FinalResults) != 2 {
		t.Fatalf("expected two results, got %d", len(payload.FinalResults))
	}
	if payload.FinalResults[0].SoundScore != payload.FinalResults[1].SoundScore {
		t.Fatalf("fixture must produce equal calibrated sound scores: %.2f vs %.2f", payload.FinalResults[0].SoundScore, payload.FinalResults[1].SoundScore)
	}
	if payload.FinalResults[0].THName != "กขคงจ" {
		t.Fatalf("raw v1 score must not break an equal calibrated tie, got %q first", payload.FinalResults[0].THName)
	}
}
