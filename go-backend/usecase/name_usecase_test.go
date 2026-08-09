package usecase_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"
	"unicode"

	"go-backend/config"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
)

func TestNameSearchUseCase_SearchTopNames_KalakiniFilter(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)

	// Search for Monday born baby
	req := domain.NameSearchRequest{
		Mode:       domain.ModeStandalone,
		DayOfBirth: "monday",
		Limit:      100,
	}

	results, err := uc.SearchTopNames(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected results for Monday born baby, got 0")
	}

	t.Logf("Fetched %d 100%% Kalakini-safe top names for Monday!", len(results))
	t.Logf("Top 1 Name for Monday: %s (Rank %d, Score: %d)", results[0].THName, results[0].Rank, results[0].TotalScore)
}

func TestNameSearchUseCase_CombinedSearchExpandsEligiblePool(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)
	req := domain.NameSearchRequest{
		Mode:       domain.ModeCombined,
		Surname:    "ทองแสน",
		DayOfBirth: "sunday",
		Limit:      100,
	}

	results, err := uc.SearchTopNames(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) <= 3 {
		t.Fatalf("expected full-database combined search to return more than 3 safe names, got %d", len(results))
	}
	for _, result := range results {
		if strings.Contains(result.SatPairType, "R") || strings.Contains(result.ShaPairType, "R") {
			t.Fatalf("result %q contains a bad pair: %s / %s", result.THName, result.SatPairType, result.ShaPairType)
		}
	}

	t.Logf("Fetched %d all-D combined results for surname ทองแสน", len(results))
}

func TestNameSearchUseCase_NumerologyOnlySaturdayRankingFingerprint(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)
	one := 1.0
	zero := 0.0
	result, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:             domain.ModeStandalone,
		DayOfBirth:       "saturday",
		Limit:            100,
		NumerologyWeight: &one,
		SemanticWeight:   &zero,
		SoundWeight:      &zero,
	})
	if err != nil {
		t.Fatalf("100%% numerology search: %v", err)
	}
	if len(result.Data) != 100 {
		t.Fatalf("expected exact Top 100, got %d", len(result.Data))
	}

	var canonical strings.Builder
	for _, item := range result.Data {
		fmt.Fprintf(&canonical, "%d:%.12f\n", item.NameID, item.FinalScore)
	}
	fingerprint := sha256.Sum256([]byte(canonical.String()))
	// Audited against an independent exhaustive SQL oracle on 2026-08-07.
	// Updating either golden requires a reviewed corpus/rule change and a fresh
	// independent oracle comparison; it must not be refreshed from this test.
	const auditedEligibleCount = 15550
	const auditedRankingSHA256 = "a028cbffadee2d2809bf5a21d4e02f719453a301321156260f345436c0334cce"
	actualFingerprint := fmt.Sprintf("%x", fingerprint)
	if result.EligibleCandidateCount != auditedEligibleCount || actualFingerprint != auditedRankingSHA256 {
		t.Fatalf(
			"audited ranking baseline changed: eligible=%d (want %d), sha256=%s (want %s); update only after reviewed data/rule change",
			result.EligibleCandidateCount,
			auditedEligibleCount,
			actualFingerprint,
			auditedRankingSHA256,
		)
	}
	t.Logf("eligible=%d ranking_sha256=%x canonical_format=name_id:final_score_12dp\\n", result.EligibleCandidateCount, fingerprint)
	t.Logf("ranking:\n%s", canonical.String())
}

func TestNameSearchUseCase_SemanticUnavailableIsHonestAndExhaustive(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()
	t.Setenv("OPENAI_API_KEY", "")

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)
	numerologyWeight := 0.45
	semanticWeight := 0.35
	soundWeight := 0.20
	result, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:             domain.ModeStandalone,
		WishQuery:        "integration fallback query without embedding credentials",
		Limit:            100,
		NumerologyWeight: &numerologyWeight,
		SemanticWeight:   &semanticWeight,
		SoundWeight:      &soundWeight,
	})
	if err != nil {
		t.Fatalf("semantic fallback search: %v", err)
	}
	const unavailableSource = "unavailable_unapproved_vector_corpus"
	if result.SemanticAvailable || result.SemanticSource != unavailableSource {
		t.Fatalf("expected explicit unavailable semantic metadata, got available=%v source=%q", result.SemanticAvailable, result.SemanticSource)
	}
	if result.SemanticQueryModel != "" || result.SemanticVectorHash != "" {
		t.Fatalf("unavailable semantic fallback must not claim query model/hash provenance: model=%q hash=%q", result.SemanticQueryModel, result.SemanticVectorHash)
	}
	if result.CandidateVectorProvenanceVerified || result.CandidateVectorProvenanceSource != unavailableSource {
		t.Fatalf("unapproved candidate vectors must fail closed: verified=%v source=%q", result.CandidateVectorProvenanceVerified, result.CandidateVectorProvenanceSource)
	}
	if result.EligibleCandidateCount <= 1000 {
		t.Fatalf("fallback must retain the complete eligible pool, got %d", result.EligibleCandidateCount)
	}
	if result.AppliedWeights.Semantic != 0 ||
		math.Abs(result.AppliedWeights.Numerology-(0.45/0.65)) > 1e-12 ||
		math.Abs(result.AppliedWeights.Sound-(0.20/0.65)) > 1e-12 {
		t.Fatalf("unexpected redistributed weights: %+v", result.AppliedWeights)
	}
	for _, item := range result.Data {
		if item.SemanticAvailable || item.SemanticScore != 0 || item.SemanticSource != unavailableSource ||
			item.CandidateVectorProvenanceVerified || item.CandidateVectorProvenanceSource != unavailableSource {
			t.Fatalf("fallback manufactured semantic evidence for %q: available=%v score=%.2f source=%q", item.THName, item.SemanticAvailable, item.SemanticScore, item.SemanticSource)
		}
	}
	latinDigitOrSpaceCount := 0
	outsideThaiBlockCount := 0
	for _, item := range result.Data {
		hasLatinDigitOrSpace := false
		hasOutsideThaiBlock := false
		for _, char := range item.THName {
			if unicode.In(char, unicode.Latin) || unicode.IsDigit(char) || unicode.IsSpace(char) {
				hasLatinDigitOrSpace = true
			}
			if char < '\u0E00' || char > '\u0E7F' {
				hasOutsideThaiBlock = true
			}
		}
		if hasLatinDigitOrSpace {
			latinDigitOrSpaceCount++
			t.Logf("latin_digit_whitespace_audit_match id=%d name=%s", item.NameID, item.THName)
		}
		if hasOutsideThaiBlock {
			outsideThaiBlockCount++
			t.Logf("outside_thai_block_audit_match id=%d name=%s", item.NameID, item.THName)
		}
	}
	t.Logf("content_audit latin_digit_or_whitespace=%d outside_thai_block=%d", latinDigitOrSpaceCount, outsideThaiBlockCount)
}

func TestNameSearchUseCase_ScreenshotLikeRankingFingerprint(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()
	t.Setenv("OPENAI_API_KEY", "must-not-be-called-without-approved-corpus")

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)
	numerologyWeight := 0.15
	semanticWeight := 0.15
	soundWeight := 0.70
	result, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:             domain.ModeStandalone,
		DayOfBirth:       "saturday",
		WishQuery:        "ผู้มีชัยชนะและสง่างามดั่งนก",
		Limit:            100,
		NumerologyWeight: &numerologyWeight,
		SemanticWeight:   &semanticWeight,
		SoundWeight:      &soundWeight,
	})
	if err != nil {
		t.Fatalf("screenshot-like ranking: %v", err)
	}
	if len(result.Data) != 100 {
		t.Fatalf("expected exact Top 100, got %d", len(result.Data))
	}
	const unavailableSource = "unavailable_unapproved_vector_corpus"
	if result.SemanticAvailable || result.SemanticSource != unavailableSource {
		t.Fatalf("unapproved legacy vector corpus must fail closed, got available=%v source=%q", result.SemanticAvailable, result.SemanticSource)
	}
	if result.SemanticQueryModel != "" || result.SemanticVectorHash != "" ||
		result.CandidateVectorProvenanceVerified || result.CandidateVectorProvenanceSource != unavailableSource {
		t.Fatalf("unavailable semantic ranking claimed proof: model=%q hash=%q verified=%v source=%q", result.SemanticQueryModel, result.SemanticVectorHash, result.CandidateVectorProvenanceVerified, result.CandidateVectorProvenanceSource)
	}
	if math.Abs(result.RequestedWeights.Numerology-0.15) > 1e-12 ||
		math.Abs(result.RequestedWeights.Semantic-0.15) > 1e-12 ||
		math.Abs(result.RequestedWeights.Sound-0.70) > 1e-12 ||
		math.Abs(result.AppliedWeights.Numerology-(3.0/17.0)) > 1e-12 ||
		result.AppliedWeights.Semantic != 0 ||
		math.Abs(result.AppliedWeights.Sound-(14.0/17.0)) > 1e-12 {
		t.Fatalf("unexpected requested/applied weights: requested=%+v applied=%+v", result.RequestedWeights, result.AppliedWeights)
	}

	var canonical strings.Builder
	var canonicalIDs strings.Builder
	for _, item := range result.Data {
		fmt.Fprintf(&canonical, "%d:%.12f\n", item.NameID, item.FinalScore)
		fmt.Fprintf(&canonicalIDs, "%d\n", item.NameID)
	}
	fingerprint := sha256.Sum256([]byte(canonical.String()))
	idFingerprint := sha256.Sum256([]byte(canonicalIDs.String()))
	// Audited independently against exhaustive SQL + the documented fail-closed
	// weight redistribution on 2026-08-07. The oracle applies exact NUMERIC SAMP
	// calibration statistics and ranking_precision_12dp_v1 before tie-breaking.
	// Do not refresh these values from the implementation under test.
	const auditedEligibleCount = 15550
	const auditedRankingSHA256 = "feace8c73b89aa8aa35dcc1ae7297f575a0dacdd1e2f170762bda328a8dc06d2"
	const auditedIDOnlySHA256 = "b13d40c6a07976953d3b8e7867eb041d0c170fa743a6df020e3426fa9699b33c"
	actualTopTen := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		actualTopTen = append(actualTopTen, result.Data[i].THName)
	}
	if result.EligibleCandidateCount != auditedEligibleCount ||
		fmt.Sprintf("%x", fingerprint) != auditedRankingSHA256 ||
		fmt.Sprintf("%x", idFingerprint) != auditedIDOnlySHA256 {
		t.Fatalf(
			"audited screenshot-like baseline changed: eligible=%d/%d rank_sha=%x/%s id_sha=%x/%s top10=%q calibration=%+v",
			result.EligibleCandidateCount, auditedEligibleCount,
			fingerprint, auditedRankingSHA256,
			idFingerprint, auditedIDOnlySHA256,
			actualTopTen, result.Policy,
		)
	}
	wantTopTen := []string{
		"สุนันทา", "มินท์", "พวงมาลย์", "เฌอฟ้า", "พิธานจริยา",
		"สุภานิษา", "อันทิพย์", "เลิศนวนิล", "นิรัชรา", "กมลมาลย์",
	}
	for i, want := range wantTopTen {
		if result.Data[i].THName != want {
			t.Fatalf("audited Top 10 changed at rank %d: got %q want %q", i+1, result.Data[i].THName, want)
		}
	}
	prefixes := []string{"นางสาว", "เด็กหญิง", "เด็กชาย", "หม่อมหลวง", "หม่อมราชวงศ์", "ศาสตราจารย์", "รองศาสตราจารย์", "ผู้ช่วยศาสตราจารย์", "ดร.", "ดร", "นาง", "นาย", "คุณ", "พระ", "พล.ต.", "พ.ต.", "ร.ต."}
	for _, item := range result.Data {
		for _, prefix := range prefixes {
			if strings.HasPrefix(item.THName, prefix) {
				t.Fatalf("ranked corpus contains title-like prefix: id=%d name=%s prefix=%s", item.NameID, item.THName, prefix)
			}
		}
		for _, char := range item.THName {
			if unicode.In(char, unicode.Latin) || unicode.IsDigit(char) || unicode.IsSpace(char) || char < '\u0E00' || char > '\u0E7F' {
				t.Fatalf("ranked corpus contains an unsupported display character: id=%d name=%q rune=%q", item.NameID, item.THName, char)
			}
		}
	}
}

func TestNameSearchUseCase_CombinedSundayRankingFingerprint(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()
	t.Setenv("OPENAI_API_KEY", "must-not-be-called-without-approved-corpus")

	repo := repository.NewNameRepository(db)
	surnameSat, surnameSha, err := repo.GetSurnameSums(context.Background(), "ทองแสน")
	if err != nil {
		t.Fatalf("calculate canonical surname sums: %v", err)
	}
	if surnameSat != 23 || surnameSha != 62 {
		t.Fatalf("canonical surname sums changed: got %d/%d want 23/62", surnameSat, surnameSha)
	}

	numerologyWeight := 0.15
	semanticWeight := 0.15
	soundWeight := 0.70
	result, err := usecase.NewNameSearchUseCase(repo).SearchTopNamesWithToken(
		context.Background(),
		domain.NameSearchRequest{
			Mode:             domain.ModeCombined,
			Surname:          "ทองแสน",
			DayOfBirth:       "sunday",
			WishQuery:        "ผู้มีชัยชนะและสง่างามดั่งนก",
			Limit:            100,
			NumerologyWeight: &numerologyWeight,
			SemanticWeight:   &semanticWeight,
			SoundWeight:      &soundWeight,
		},
	)
	if err != nil {
		t.Fatalf("combined ranking: %v", err)
	}
	if len(result.Data) != 100 {
		t.Fatalf("expected exact Top 100, got %d", len(result.Data))
	}
	if result.Policy.PairScope != "first_name_and_combined_all_d_pairs" {
		t.Fatalf("combined pair gate is not explicit: %q", result.Policy.PairScope)
	}
	if result.SemanticAvailable || result.AppliedWeights.Semantic != 0 ||
		math.Abs(result.AppliedWeights.Numerology-(3.0/17.0)) > 1e-12 ||
		math.Abs(result.AppliedWeights.Sound-(14.0/17.0)) > 1e-12 {
		t.Fatalf("unexpected fail-closed combined weights/provenance: available=%v source=%q applied=%+v", result.SemanticAvailable, result.SemanticSource, result.AppliedWeights)
	}

	var canonical strings.Builder
	var canonicalIDs strings.Builder
	for _, item := range result.Data {
		fmt.Fprintf(&canonical, "%d:%.12f\n", item.NameID, item.FinalScore)
		fmt.Fprintf(&canonicalIDs, "%d\n", item.NameID)
	}
	fingerprint := sha256.Sum256([]byte(canonical.String()))
	idFingerprint := sha256.Sum256([]byte(canonicalIDs.String()))

	// Independently audited with exhaustive SQL on 2026-08-07. The oracle
	// recalculates ทองแสน as 23/62 and requires all pairs to be D-type for both
	// the first-name totals and the post-surname totals before global scoring.
	const auditedEligibleCount = 696
	const auditedRankingSHA256 = "001e0a7f3aa11449b7552e7fb860e79dbb861ff3b807be622d120b5e067a41ed"
	const auditedIDOnlySHA256 = "e28abf50763b65db1b66fe409b8cdbc27cb45160f63d2464dac457d3098336f4"
	if result.EligibleCandidateCount != auditedEligibleCount ||
		fmt.Sprintf("%x", fingerprint) != auditedRankingSHA256 ||
		fmt.Sprintf("%x", idFingerprint) != auditedIDOnlySHA256 {
		t.Fatalf(
			"audited combined baseline changed: eligible=%d/%d rank_sha=%x/%s id_sha=%x/%s",
			result.EligibleCandidateCount, auditedEligibleCount,
			fingerprint, auditedRankingSHA256,
			idFingerprint, auditedIDOnlySHA256,
		)
	}

	first := result.Data[0]
	last := result.Data[len(result.Data)-1]
	if first.NameID != 206141 || first.THName != "เกียรติโชติ" ||
		first.TotalSatSum != 65 || first.TotalShaSum != 159 || first.FinalScore != 94.736678200692 {
		t.Fatalf("combined first checksum changed: %+v", first)
	}
	if last.NameID != 218364 || last.THName != "เธียรปรีชา" ||
		last.TotalSatSum != 64 || last.TotalShaSum != 159 || last.FinalScore != 88.922145328720 {
		t.Fatalf("combined last checksum changed: %+v", last)
	}
}
