package usecase

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go-backend/domain"
)

func TestCandidatePoolCacheExpiresAndEvictsOldestEntry(t *testing.T) {
	currentTime := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	tokens := []string{"first", "second", "third"}
	tokenIndex := 0
	cache := newCandidatePoolCache(2, time.Minute)
	cache.now = func() time.Time { return currentTime }
	cache.newToken = func() (string, error) {
		token := tokens[tokenIndex]
		tokenIndex++
		return token, nil
	}

	first, err := cache.Put(cachedCandidatePool{})
	if err != nil {
		t.Fatalf("put first pool: %v", err)
	}
	currentTime = currentTime.Add(time.Second)
	second, err := cache.Put(cachedCandidatePool{})
	if err != nil {
		t.Fatalf("put second pool: %v", err)
	}
	currentTime = currentTime.Add(time.Second)
	third, err := cache.Put(cachedCandidatePool{})
	if err != nil {
		t.Fatalf("put third pool: %v", err)
	}

	if _, ok := cache.Get(first); ok {
		t.Fatal("expected oldest pool to be evicted at capacity")
	}
	if _, ok := cache.Get(second); !ok {
		t.Fatal("expected second pool to remain cached")
	}
	if _, ok := cache.Get(third); !ok {
		t.Fatal("expected newest pool to remain cached")
	}

	currentTime = currentTime.Add(time.Minute)
	if _, ok := cache.Get(second); ok {
		t.Fatal("expected pool to expire after its TTL")
	}
}

func TestCandidatePoolCacheDefensivelyCopiesMutableData(t *testing.T) {
	cache := newCandidatePoolCache(1, time.Minute)
	cache.newToken = func() (string, error) { return "token", nil }
	weight := 0.5
	pool := cachedCandidatePool{
		request: domain.NameSearchRequest{NumerologyWeight: &weight},
		goodNumbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
		},
		candidates: []domain.NameResultItem{{NameID: 1, THName: "เดิม", RankingReasons: []string{"เดิม"}}},
	}

	token, err := cache.Put(pool)
	if err != nil {
		t.Fatalf("put pool: %v", err)
	}
	weight = 1
	pool.goodNumbers["24"] = domain.NumberMeaning{PairType: "R5"}
	pool.candidates[0].THName = "เปลี่ยน"

	firstRead, ok := cache.Get(token)
	if !ok {
		t.Fatal("expected cached pool")
	}
	if *firstRead.request.NumerologyWeight != 0.5 || firstRead.goodNumbers["24"].PairType != "D10" || firstRead.candidates[0].THName != "เดิม" {
		t.Fatal("cache retained a mutable reference from its caller")
	}

	*firstRead.request.NumerologyWeight = 0
	firstRead.candidates[0].RankingReasons[0] = "เปลี่ยน"
	secondRead, ok := cache.Get(token)
	if !ok {
		t.Fatal("expected cached pool on second read")
	}
	if *secondRead.request.NumerologyWeight != 0.5 || secondRead.candidates[0].RankingReasons[0] != "เดิม" {
		t.Fatal("cache returned shared mutable data")
	}
}

func TestSearchTokenReranksPreparedPoolWithOnlyTopMeaningLookup(t *testing.T) {
	repo := &countingNameRepository{
		numbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D10", PairPoint: 80},
			"45": {PairType: "D10", PairPoint: 85},
			"99": {PairType: "D10", PairPoint: 100},
		},
		candidates: []domain.NameResultItem{
			{NameID: 1, THName: "เสียงเด่น", SatSum: 24, ShaSum: 24, EuphonyScore: 100, PronunciationEase: 100, RhythmScore: 100, PhoneticVersion: "v1", SemanticAvailable: true, SemanticSource: "embedding", SemanticQueryModel: "text-embedding-3-small", SemanticVectorHash: "proof-hash"},
			{NameID: 2, THName: "เลขเด่น", SatSum: 45, ShaSum: 45, PhoneticVersion: "v1", SemanticAvailable: true, SemanticSource: "embedding", SemanticQueryModel: "text-embedding-3-small", SemanticVectorHash: "proof-hash"},
		},
		semanticCorpus: domain.SemanticCorpusProvenance{
			SnapshotID:             7,
			Approved:               true,
			Source:                 "approved_snapshot",
			EmbeddingModel:         domain.SemanticEmbeddingModel,
			InputTemplateVersion:   domain.SemanticInputTemplateVersion,
			VectorDimensions:       domain.SemanticVectorDimensions,
			CorpusFilterVersion:    domain.CorpusFilterVersion,
			CorpusVectorSHA256:     strings.Repeat("a", 64),
			CorpusMembershipSHA256: strings.Repeat("b", 64),
			EvaluationVersion:      "held_out_v1",
		},
		soundCalibration: domain.SoundCalibrationProfile{
			FormulaVersion:   domain.SoundFormulaVersion,
			Source:           "live_repository_distribution_provisional",
			ReferenceVersion: "v1",
			ReferenceMean:    80,
			ReferenceStdDev:  10,
			Versions: map[string]domain.SoundCalibrationStats{
				"v1": {Count: 2, Mean: 80, StdDev: 10},
			},
		},
	}
	uc := NewNameSearchUseCase(repo)
	one := 1.0
	zero := 0.0

	initial, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:             domain.ModeStandalone,
		WishQuery:        "test provenance",
		Limit:            2,
		NumerologyWeight: &one,
		SemanticWeight:   &zero,
		SoundWeight:      &zero,
	})
	if err != nil {
		t.Fatalf("initial search: %v", err)
	}
	if initial.CandidateToken == "" {
		t.Fatal("expected an opaque candidate token")
	}
	if initial.Policy.PairScope != "first_name_all_known_pairs_scored" ||
		initial.Policy.KalakiniScope != "not_applied" ||
		initial.Policy.SoundScope != "first_name" ||
		initial.Policy.NumerologyRuleVersion != domain.NumerologyRuleVersion ||
		initial.Policy.SoundFormulaVersion != domain.SoundFormulaVersion ||
		initial.Policy.RankingPrecisionVersion != domain.RankingPrecisionVersion ||
		initial.Policy.SoundCalibrationSource != "live_repository_distribution_provisional" ||
		initial.Policy.SoundCalibrationReference != "v1" ||
		initial.Policy.SoundCalibrationApproved ||
		initial.Policy.CorpusFilterVersion != domain.CorpusFilterVersion ||
		!initial.Policy.SemanticCorpusApproved ||
		initial.Policy.SemanticCorpusSource != "approved_snapshot" {
		t.Fatalf("unexpected explicit ranking policy: %+v", initial.Policy)
	}
	if len(initial.Data) != 2 || initial.Data[0].THName != "เลขเด่น" {
		t.Fatalf("expected numerology-first initial order, got %+v", initial.Data)
	}
	repositoryCallsAfterSearch := repo.callCount()

	reranked, err := uc.RerankTopNames(context.Background(), domain.NameRerankRequest{
		CandidateToken:   initial.CandidateToken,
		NumerologyWeight: &zero,
		SemanticWeight:   &zero,
		SoundWeight:      &one,
		Limit:            2,
	})
	if err != nil {
		t.Fatalf("rerank: %v", err)
	}
	if len(reranked.Data) != 2 || reranked.Data[0].THName != "เสียงเด่น" {
		t.Fatalf("expected sound-first reranked order, got %+v", reranked.Data)
	}
	if reranked.CandidateToken != initial.CandidateToken {
		t.Fatal("rerank should preserve the candidate token")
	}
	if reranked.SemanticQueryModel != "text-embedding-3-small" || reranked.SemanticVectorHash != "proof-hash" ||
		reranked.Data[0].SemanticVectorHash != "proof-hash" {
		t.Fatalf("rerank lost cached semantic proof provenance: model=%q hash=%q item=%q", reranked.SemanticQueryModel, reranked.SemanticVectorHash, reranked.Data[0].SemanticVectorHash)
	}
	if !reranked.CandidateVectorProvenanceVerified || reranked.CandidateVectorProvenanceSource != "approved_snapshot" {
		t.Fatalf("rerank lost cached candidate-vector provenance: verified=%v source=%q", reranked.CandidateVectorProvenanceVerified, reranked.CandidateVectorProvenanceSource)
	}
	if calls := repo.callCount(); calls != repositoryCallsAfterSearch+1 {
		t.Fatalf("rerank should make only one Top-N meaning lookup: before=%d after=%d", repositoryCallsAfterSearch, calls)
	}
}

func TestRerankRequestValidationAndUnknownToken(t *testing.T) {
	uc := NewNameSearchUseCase(&countingNameRepository{})
	validToken, err := generateCandidateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	one := 1.0
	zero := 0.0
	tooLarge := 1.1

	tests := []struct {
		name string
		req  domain.NameRerankRequest
	}{
		{name: "missing token", req: domain.NameRerankRequest{NumerologyWeight: &one, SemanticWeight: &zero, SoundWeight: &zero}},
		{name: "missing weight", req: domain.NameRerankRequest{CandidateToken: validToken, NumerologyWeight: &one, SemanticWeight: &zero}},
		{name: "out of range weight", req: domain.NameRerankRequest{CandidateToken: validToken, NumerologyWeight: &tooLarge, SemanticWeight: &zero, SoundWeight: &zero}},
		{name: "zero total", req: domain.NameRerankRequest{CandidateToken: validToken, NumerologyWeight: &zero, SemanticWeight: &zero, SoundWeight: &zero}},
		{name: "large limit", req: domain.NameRerankRequest{CandidateToken: validToken, NumerologyWeight: &one, SemanticWeight: &zero, SoundWeight: &zero, Limit: 101}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.RerankTopNames(context.Background(), tt.req)
			if !errors.Is(err, domain.ErrInvalidRerankRequest) {
				t.Fatalf("expected invalid rerank error, got %v", err)
			}
		})
	}

	_, err = uc.RerankTopNames(context.Background(), domain.NameRerankRequest{
		CandidateToken:   validToken,
		NumerologyWeight: &one,
		SemanticWeight:   &zero,
		SoundWeight:      &zero,
		Limit:            10,
	})
	if !errors.Is(err, domain.ErrCandidateTokenNotFound) {
		t.Fatalf("expected missing-token error, got %v", err)
	}
}

func TestNumerologyOnlySearchRanksTheCompletePoolBeforeTakingTopOneHundred(t *testing.T) {
	candidates := make([]domain.NameResultItem, 0, 1001)
	for i := 1; i <= 1000; i++ {
		candidates = append(candidates, domain.NameResultItem{
			NameID: i,
			THName: "คะแนนรอง",
			SatSum: 24,
			ShaSum: 24,
		})
	}
	candidates = append(candidates, domain.NameResultItem{
		NameID: 1001,
		THName: "คะแนนสูงสุดอยู่นอกหนึ่งพันแถวแรก",
		SatSum: 99,
		ShaSum: 99,
	})

	repo := &countingNameRepository{
		numbers: map[string]domain.NumberMeaning{
			"24": {PairType: "D8", PairPoint: 80},
			"99": {PairType: "D10", PairPoint: 100},
		},
		candidates: candidates,
	}
	uc := NewNameSearchUseCase(repo)
	one := 1.0
	zero := 0.0

	result, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:             domain.ModeStandalone,
		Limit:            100,
		NumerologyWeight: &one,
		SemanticWeight:   &zero,
		SoundWeight:      &zero,
	})
	if err != nil {
		t.Fatalf("search complete pool: %v", err)
	}
	if len(result.Data) != 100 {
		t.Fatalf("expected exact top 100, got %d", len(result.Data))
	}
	if result.Data[0].NameID != 1001 {
		t.Fatalf("expected the true numerical maximum first, got id=%d score=%.2f", result.Data[0].NameID, result.Data[0].FinalScore)
	}
	if result.EligibleCandidateCount != 1001 {
		t.Fatalf("expected all 1001 eligible candidates to be scored, got %d", result.EligibleCandidateCount)
	}
	repo.mu.Lock()
	lastLimit := repo.lastCandidateLimit
	repo.mu.Unlock()
	if lastLimit != 0 {
		t.Fatalf("candidate query must be exhaustive before scoring; got repository limit %d", lastLimit)
	}
}

func TestInitialSearchRequestValidation(t *testing.T) {
	one := 1.0
	zero := 0.0
	tests := []struct {
		name string
		req  domain.NameSearchRequest
	}{
		{name: "unsupported mode", req: domain.NameSearchRequest{Mode: "mystery"}},
		{name: "unsupported day", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, DayOfBirth: "someday"}},
		{name: "blank combined surname", req: domain.NameSearchRequest{Mode: domain.ModeCombined, Surname: "  "}},
		{name: "semantic without wish", req: domain.NameSearchRequest{Mode: domain.ModeSemantic}},
		{name: "name intent without seed", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, IsNameIntent: true}},
		{name: "negative limit", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, Limit: -1}},
		{name: "limit above one hundred", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, Limit: 101}},
		{name: "partial three-way weights", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, NumerologyWeight: &one, SemanticWeight: &zero}},
		{name: "zero three-way weights", req: domain.NameSearchRequest{Mode: domain.ModeStandalone, NumerologyWeight: &zero, SemanticWeight: &zero, SoundWeight: &zero}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewNameSearchUseCase(&countingNameRepository{})
			_, err := uc.SearchTopNamesWithToken(context.Background(), tt.req)
			if !errors.Is(err, domain.ErrInvalidSearchRequest) {
				t.Fatalf("expected invalid search request, got %v", err)
			}
		})
	}
}

func TestCombinedSearchRejectsUnmappedSurnameCharacters(t *testing.T) {
	repo := &countingNameRepository{
		surnameErr: errors.Join(domain.ErrUnmappedCharacters, errors.New("unmapped rune U+1F642")),
	}
	uc := NewNameSearchUseCase(repo)

	_, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:    domain.ModeCombined,
		Surname: "🙂",
	})
	if !errors.Is(err, domain.ErrInvalidSearchRequest) {
		t.Fatalf("expected invalid search request for unmapped surname, got %v", err)
	}
}

func TestDefaultCandidateCacheCapacityStaysMemoryBounded(t *testing.T) {
	if defaultCandidateCacheCapacity > 8 {
		t.Fatalf("complete candidate pools are large; expected capacity at most 8, got %d", defaultCandidateCacheCapacity)
	}
}

func TestWishCandidateQueriesAreExhaustiveBeforeRanking(t *testing.T) {
	tests := []struct {
		name         string
		isNameIntent bool
		readLimit    func(*countingNameRepository) int
	}{
		{
			name:         "semantic embedding search",
			isNameIntent: false,
			readLimit:    func(repo *countingNameRepository) int { return repo.lastSemanticLimit },
		},
		{
			name:         "trigram name-intent search",
			isNameIntent: true,
			readLimit:    func(repo *countingNameRepository) int { return repo.lastSimilarLimit },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &countingNameRepository{
				numbers: map[string]domain.NumberMeaning{
					"24": {PairType: "D10", PairPoint: 80},
				},
				candidates: []domain.NameResultItem{{
					NameID:            1,
					THName:            "ทดสอบ",
					SatSum:            24,
					ShaSum:            24,
					SemanticAvailable: true,
					SemanticSource:    "test",
				}},
			}
			uc := NewNameSearchUseCase(repo)
			_, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
				Mode:         domain.ModeStandalone,
				WishQuery:    "ความหมายทดสอบ",
				IsNameIntent: tt.isNameIntent,
				Limit:        1,
			})
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			repo.mu.Lock()
			limit := tt.readLimit(repo)
			repo.mu.Unlock()
			if limit != 0 {
				t.Fatalf("expected exhaustive repository limit 0, got %d", limit)
			}
		})
	}
}

func TestPrototypeNameAndMeaningQueryRemainSeparate(t *testing.T) {
	repo := &countingNameRepository{
		numbers:    map[string]domain.NumberMeaning{"24": {PairType: "D10", PairPoint: 80}},
		candidates: []domain.NameResultItem{{NameID: 1, THName: "ณเดช", SatSum: 24, ShaSum: 24}},
	}
	uc := NewNameSearchUseCase(repo)
	_, err := uc.SearchTopNamesWithToken(context.Background(), domain.NameSearchRequest{
		Mode:          domain.ModeStandalone,
		PrototypeName: "ณเดชน์",
		WishQuery:     "ผู้มีชัยชนะและสง่างาม",
		IsNameIntent:  true,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.lastPrototypeName != "ณเดชน์" || repo.lastMeaningQuery != "ผู้มีชัยชนะและสง่างาม" {
		t.Fatalf("prototype and meaning were mixed: prototype=%q meaning=%q", repo.lastPrototypeName, repo.lastMeaningQuery)
	}
}

type countingNameRepository struct {
	mu                 sync.Mutex
	calls              int
	numbers            map[string]domain.NumberMeaning
	candidates         []domain.NameResultItem
	soundCalibration   domain.SoundCalibrationProfile
	semanticCorpus     domain.SemanticCorpusProvenance
	surnameSatSum      int
	surnameShaSum      int
	surnameErr         error
	maxSatSum          int
	maxShaSum          int
	lastCandidateLimit int
	lastSemanticLimit  int
	lastSimilarLimit   int
	lastPrototypeName  string
	lastMeaningQuery   string
}

func (r *countingNameRepository) recordCall() {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
}

func (r *countingNameRepository) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *countingNameRepository) GetGoodNumbersMap(context.Context) (map[string]domain.NumberMeaning, error) {
	r.recordCall()
	return r.numbers, nil
}

func (r *countingNameRepository) GetAllNumbersMap(context.Context) (map[string]domain.NumberMeaning, error) {
	r.recordCall()
	return r.numbers, nil
}

func (r *countingNameRepository) GetSurnameSums(context.Context, string) (int, int, error) {
	r.recordCall()
	return r.surnameSatSum, r.surnameShaSum, r.surnameErr
}

func (r *countingNameRepository) GetNameSumMaximums(context.Context) (int, int, error) {
	r.recordCall()
	maxSatSum := r.maxSatSum
	maxShaSum := r.maxShaSum
	if maxSatSum == 0 {
		maxSatSum = 100
	}
	if maxShaSum == 0 {
		maxShaSum = 100
	}
	return maxSatSum, maxShaSum, nil
}

func (r *countingNameRepository) GetSoundCalibrationProfile(context.Context) (domain.SoundCalibrationProfile, error) {
	r.recordCall()
	return cloneSoundCalibrationProfile(r.soundCalibration), nil
}

func (r *countingNameRepository) GetSemanticCorpusProvenance(context.Context) (domain.SemanticCorpusProvenance, error) {
	r.recordCall()
	return r.semanticCorpus, nil
}

func (r *countingNameRepository) GetSumsByTHName(context.Context, string) (int, int, string, error) {
	r.recordCall()
	return 0, 0, "", nil
}

func (r *countingNameRepository) GetMeaningsByNameIDs(_ context.Context, nameIDs []int) (map[int]string, error) {
	r.recordCall()
	wanted := make(map[int]bool, len(nameIDs))
	for _, nameID := range nameIDs {
		wanted[nameID] = true
	}
	result := make(map[int]string, len(nameIDs))
	for _, item := range r.candidates {
		if wanted[item.NameID] {
			result[item.NameID] = item.Meaning
		}
	}
	return result, nil
}

func TestEnrichResultMeaningsKeepsThaiFullNameOrder(t *testing.T) {
	repo := &countingNameRepository{candidates: []domain.NameResultItem{{NameID: 1, Meaning: "ความหมายนามสกุล"}}}
	uc := &nameSearchUseCase{repo: repo}
	results := []domain.NameResultItem{{NameID: 1}}
	req := domain.NameSearchRequest{Mode: domain.ModeCombined, SearchTarget: domain.TargetSurname}
	if err := uc.enrichResultMeanings(context.Background(), results, req, "ความหมายชื่อตัว"); err != nil {
		t.Fatalf("enrich result meanings: %v", err)
	}
	if results[0].Meaning != "ความหมายชื่อตัว ความหมายนามสกุล" {
		t.Fatalf("full-name meaning order is wrong: %q", results[0].Meaning)
	}
}

func (r *countingNameRepository) SearchCandidateNames(_ context.Context, _ []int, _ []int, _ string, _ string, limit int) ([]domain.NameResultItem, error) {
	r.mu.Lock()
	r.calls++
	r.lastCandidateLimit = limit
	r.mu.Unlock()
	return cloneNameResults(r.candidates), nil
}

func (r *countingNameRepository) SearchSemanticWishNames(_ context.Context, _ []int, _ []int, _ string, _ string, limit int) ([]domain.NameResultItem, error) {
	r.mu.Lock()
	r.calls++
	r.lastSemanticLimit = limit
	r.mu.Unlock()
	return cloneNameResults(r.candidates), nil
}

func (r *countingNameRepository) SearchSimilarNames(_ context.Context, _ []int, _ []int, _ string, _ string, limit int) ([]domain.NameResultItem, error) {
	r.mu.Lock()
	r.calls++
	r.lastSimilarLimit = limit
	r.mu.Unlock()
	return cloneNameResults(r.candidates), nil
}

func (r *countingNameRepository) GetMeaningVectorByTHName(context.Context, string) (string, string, error) {
	r.recordCall()
	return "", "", nil
}

func (r *countingNameRepository) SearchSemanticAndSimilarNames(_ context.Context, _ []int, _ []int, prototypeName string, meaningQuery string, _ string, limit int) ([]domain.NameResultItem, error) {
	r.mu.Lock()
	r.calls++
	r.lastSimilarLimit = limit // Reuse similar limit tracker for test purposes
	r.lastPrototypeName = prototypeName
	r.lastMeaningQuery = meaningQuery
	r.mu.Unlock()
	return cloneNameResults(r.candidates), nil
}

func (r *countingNameRepository) GetSampleNames(context.Context) ([]domain.Celebrity, error) {
	r.recordCall()
	return nil, nil
}

func (r *countingNameRepository) GetRandomMeanings(context.Context, int) ([]string, error) {
	r.recordCall()
	return nil, nil
}

func (r *countingNameRepository) SuggestNames(_ context.Context, _ string, _ int) ([]domain.NameResultItem, error) {
	r.recordCall()
	return cloneNameResults(r.candidates), nil
}
