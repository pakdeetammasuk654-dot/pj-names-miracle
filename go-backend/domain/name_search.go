package domain

import (
	"context"
	"errors"
	"strings"
)

type SearchMode string
type SearchTarget string

const (
	ModeCombined                 SearchMode   = "combined"   // ชื่อ + นามสกุล
	ModeStandalone               SearchMode   = "standalone" // เฉพาะชื่อ
	ModeSemantic                 SearchMode   = "semantic"   // ค้นหาด้วย AI Semantic Wish Search
	TargetFirstName              SearchTarget = "first_name"
	TargetSurname                SearchTarget = "surname"
	NumerologyRuleVersion                     = "thai_sum_v1_dictionary"
	SoundFormulaVersion                       = "sound_45_35_20_v1"
	RankingPrecisionVersion                   = "ranking_precision_12dp_v1"
	MissingPhoneticVersion                    = "<missing>"
	CorpusFilterVersion                       = "title_and_dictionary_coverage_v2"
	SemanticEmbeddingModel                    = "text-embedding-3-small"
	SemanticInputTemplateVersion              = "meaning_prefix_th_v1"
	SemanticVectorDimensions                  = 1536
	APIVersion                                = "1"
	RankingContractVersion                    = "adaptive_pairs_v1"
)

type NameSearchRequest struct {
	Mode             SearchMode   `json:"mode"`                        // "combined", "standalone", "semantic"
	SearchTarget     SearchTarget `json:"search_target,omitempty"`     // combined: candidate is first_name or surname
	Surname          string       `json:"surname,omitempty"`           // นามสกุล (ถ้าเป็น combined)
	DayOfBirth       string       `json:"day_of_birth,omitempty"`      // วันเกิด
	PrototypeName    string       `json:"prototype_name,omitempty"`    // ชื่อ/นามสกุลต้นแบบสำหรับ pg_trgm similarity
	WishQuery        string       `json:"wish_query,omitempty"`        // ความหมายของชื่อต้นแบบสำหรับ semantic ranking
	Limit            int          `json:"limit,omitempty"`             // Default 100
	NumerologyWeight *float64     `json:"numerology_weight,omitempty"` // Weight for เลขศาสตร์
	SemanticWeight   *float64     `json:"semantic_weight,omitempty"`   // Weight for meaning relevance
	SoundWeight      *float64     `json:"sound_weight,omitempty"`      // Weight for euphony and pronunciation quality
	SimilarityWeight *float64     `json:"similarity_weight,omitempty"` // Weight for name similarity (trigram/phonetic)
	IsNameIntent     bool         `json:"is_name_intent,omitempty"`    // Compatibility flag: prototype_name is present
}

// NameRerankRequest changes only the ranking weights and response size for an
// existing candidate pool. CandidateToken is intentionally opaque to clients.
type NameRerankRequest struct {
	CandidateToken   string   `json:"candidate_token"`
	NumerologyWeight *float64 `json:"numerology_weight"`
	SemanticWeight   *float64 `json:"semantic_weight"`
	SoundWeight      *float64 `json:"sound_weight"`
	SimilarityWeight *float64 `json:"similarity_weight"`
	Limit            int      `json:"limit,omitempty"`
}

// NameSearchResult carries the ranked data together with the reusable
// candidate-pool token. HTTP handlers retain the existing response envelope.
type NameSearchResult struct {
	CandidateToken                    string
	Mode                              SearchMode
	Data                              []NameResultItem
	EligibleCandidateCount            int
	RequestedWeights                  RankingWeights
	AppliedWeights                    RankingWeights
	SemanticAvailable                 bool
	SemanticSource                    string
	SemanticQueryModel                string
	SemanticModel                     string // Deprecated API alias for SemanticQueryModel.
	SemanticVectorHash                string
	CandidateVectorProvenanceVerified bool
	CandidateVectorProvenanceSource   string
	SemanticCorpus                    SemanticCorpusProvenance
	Policy                            NameSearchPolicy
}

var (
	ErrInvalidSearchRequest   = errors.New("invalid name search request")
	ErrInvalidRerankRequest   = errors.New("invalid rerank request")
	ErrCandidateTokenNotFound = errors.New("candidate token not found or expired")
	ErrUnmappedCharacters     = errors.New("text contains characters missing from numerology dictionaries")
)

// RankingWeights is the normalized formula actually used by the backend.
// RequestedWeights reflects the selected formula before an unavailable
// semantic dimension is redistributed; AppliedWeights is the final formula.
type RankingWeights struct {
	Numerology float64 `json:"numerology"`
	Semantic   float64 `json:"semantic"`
	Sound      float64 `json:"sound"`
	Similarity float64 `json:"similarity"`
}

// NameSearchPolicy makes calculation scope explicit to API clients. Combined
// mode evaluates both the candidate and combined totals with the adaptive pair
// gate selected by the slider. Kalakini and sound apply to the searched
// candidate; the fixed user-entered part is preserved as-is.
type NameSearchPolicy struct {
	RankingContractVersion        string `json:"ranking_contract_version"`
	PairScope                     string `json:"pair_scope"`
	KalakiniScope                 string `json:"kalakini_scope"`
	SoundScope                    string `json:"sound_scope"`
	NumerologyRuleVersion         string `json:"numerology_rule_version"`
	SoundFormulaVersion           string `json:"sound_formula_version"`
	RankingPrecisionVersion       string `json:"ranking_precision_version"`
	SoundCalibrationSource        string `json:"sound_calibration_source"`
	SoundCalibrationReference     string `json:"sound_calibration_reference_version"`
	SoundCalibrationApproved      bool   `json:"sound_calibration_approved"`
	SoundCalibrationAuditAt       string `json:"sound_calibration_audit_at,omitempty"`
	SoundProfileLoadedAt          string `json:"sound_profile_loaded_at,omitempty"`
	SoundEvaluationVersion        string `json:"sound_evaluation_version,omitempty"`
	SoundCorpusFingerprint        string `json:"sound_corpus_fingerprint_sha256,omitempty"`
	CorpusFilterVersion           string `json:"corpus_filter_version"`
	SemanticQueryModel            string `json:"semantic_query_model"`
	SemanticInputTemplate         string `json:"semantic_input_template_version"`
	SemanticVectorDimensions      int    `json:"semantic_vector_dimensions"`
	SemanticCorpusApproved        bool   `json:"semantic_candidate_corpus_approved"`
	SemanticCorpusSource          string `json:"semantic_candidate_corpus_source"`
	SemanticCorpusFingerprint     string `json:"semantic_candidate_corpus_sha256,omitempty"`
	SemanticMembershipFingerprint string `json:"semantic_candidate_membership_sha256,omitempty"`
	SemanticCorpusAuditAt         string `json:"semantic_candidate_corpus_audit_at,omitempty"`
	SemanticEvaluationVersion     string `json:"semantic_evaluation_version,omitempty"`
}

// SoundCalibrationStats stores one phonetic model/version distribution.
// Scoring maps every version onto a data-selected reference distribution so
// values produced by different Typhoon versions are comparable.
type SoundCalibrationStats struct {
	Count  int64
	Mean   float64
	StdDev float64
}

type SoundCalibrationProfile struct {
	ReferenceVersion        string
	ReferenceMean           float64
	ReferenceStdDev         float64
	FormulaVersion          string
	Source                  string
	Approved                bool
	SourceAuditAt           string
	LoadedAt                string
	EvaluationVersion       string
	CorpusFingerprintSHA256 string
	Versions                map[string]SoundCalibrationStats
}

// SemanticCorpusProvenance proves that stored candidate vectors were produced
// with a compatible model/template and reviewed corpus snapshot. Without an
// approved record, semantic ranking is unavailable rather than guessed.
type SemanticCorpusProvenance struct {
	SnapshotID             int64
	Approved               bool
	Source                 string
	EmbeddingModel         string
	InputTemplateVersion   string
	VectorDimensions       int
	CorpusFilterVersion    string
	CorpusRowCount         int64
	VectorizedRowCount     int64
	CorpusVectorSHA256     string
	CorpusMembershipSHA256 string
	SourceAuditAt          string
	EvaluationVersion      string
}

type NameResultItem struct {
	NameID                            int                `json:"name_id"`
	THName                            string             `json:"thname"`
	Meaning                           string             `json:"meaning,omitempty"`             // ความหมายของชื่อ
	VectorDistance                    float64            `json:"vector_distance,omitempty"`     // ระยะทาง Vector Similarity
	SimilarityDistance                float64            `json:"similarity_distance,omitempty"` // ระยะทาง pg_trgm similarity
	SatSum                            int                `json:"sat_sum"`
	ShaSum                            int                `json:"sha_sum"`
	TotalSatSum                       int                `json:"total_sat_sum,omitempty"`
	TotalShaSum                       int                `json:"total_sha_sum,omitempty"`
	SatPairsDisplay                   string             `json:"sat_pairs_display"` // ตัวเลขคู่ 00 - 99
	ShaPairsDisplay                   string             `json:"sha_pairs_display"` // ตัวเลขคู่ 00 - 99
	SatPairs                          []PairPresentation `json:"sat_pairs"`
	ShaPairs                          []PairPresentation `json:"sha_pairs"`
	CandidateSatPairs                 []PairPresentation `json:"candidate_sat_pairs,omitempty"`
	CandidateShaPairs                 []PairPresentation `json:"candidate_sha_pairs,omitempty"`
	SatPairType                       string             `json:"sat_pairtype"` // D10, D8, D5
	ShaPairType                       string             `json:"sha_pairtype"` // D10, D8, D5
	HasBadPair                        bool               `json:"has_bad_pair"`
	BadPairCount                      int                `json:"bad_pair_count"`
	PairRiskLevel                     string             `json:"pair_risk_level,omitempty"` // strict, caution_r5, caution_r7
	SatMiracleDesc                    string             `json:"sat_miracledesc"`
	ShaMiracleDesc                    string             `json:"sha_miracledesc"`
	IsPerfectCombined                 bool               `json:"is_perfect_combined"` // true = ผลรวม (ชื่อ + นามสกุล) ได้เกรด D10/D8/D5 สมบูรณ์แบบ
	TotalScore                        int                `json:"total_score"`         // คะแนนรวมสำหรับ Rank
	NumerologyScore                   float64            `json:"numerology_score"`    // คะแนนเลขศาสตร์ที่ normalize แล้ว (0-100)
	SemanticScore                     float64            `json:"semantic_score"`      // คะแนนความหมายที่ normalize แล้ว (0-100)
	SimilarityScore                   float64            `json:"similarity_score"`    // คะแนนความคล้ายชื่อ (0-100)
	SemanticAvailable                 bool               `json:"semantic_available"`
	SemanticSource                    string             `json:"semantic_source"`
	SemanticQueryModel                string             `json:"semantic_query_model,omitempty"`
	SemanticModel                     string             `json:"semantic_model,omitempty"` // Deprecated alias.
	SemanticVectorHash                string             `json:"semantic_vector_hash,omitempty"`
	CandidateVectorProvenanceVerified bool               `json:"candidate_vector_provenance_verified"`
	CandidateVectorProvenanceSource   string             `json:"candidate_vector_provenance_source"`
	SoundScore                        float64            `json:"sound_score"`     // คะแนนเสียงที่ calibrate แล้ว (0-100)
	RawSoundScore                     float64            `json:"raw_sound_score"` // คะแนนจากองค์ประกอบดิบก่อน calibrate
	SoundScope                        string             `json:"sound_scope"`     // first_name or surname candidate
	PhoneticVersion                   string             `json:"phonetic_version"`
	SoundCalibration                  string             `json:"sound_calibration"`
	PronunciationEase                 int                `json:"pronunciation_ease"` // ความง่ายในการออกเสียง
	EuphonyScore                      int                `json:"euphony_score"`      // ความไพเราะ
	RhythmScore                       int                `json:"rhythm_score"`       // จังหวะของชื่อ
	PerfectBonus                      float64            `json:"perfect_bonus"`      // Deprecated; retained as zero for response compatibility
	FinalScore                        float64            `json:"final_score"`        // คะแนนรวมแบบ Hybrid ถ่วงน้ำหนัก
	RankingReasons                    []string           `json:"ranking_reasons"`    // เหตุผลสั้น ๆ ที่อธิบายอันดับ
	Rank                              int                `json:"rank"`
}

type PairPresentation struct {
	Value  string `json:"value"`
	Type   string `json:"type"`
	IsGood bool   `json:"is_good"`
}

type NameSearchPayload struct {
	Request                           NameSearchRequest
	SurnameSatSum                     int
	SurnameShaSum                     int
	SurnameMeaning                    string
	GoodNumbers                       map[string]NumberMeaning // exact dictionary key -> NumberMeaning
	SoundCalibration                  SoundCalibrationProfile
	Candidates                        []NameResultItem
	FinalResults                      []NameResultItem
	RequestedWeights                  RankingWeights
	AppliedWeights                    RankingWeights
	SemanticAvailable                 bool
	SemanticSource                    string
	SemanticQueryModel                string
	SemanticVectorHash                string
	CandidateVectorProvenanceVerified bool
	CandidateVectorProvenanceSource   string
	SemanticCorpus                    SemanticCorpusProvenance
}

// CombineMeanings joins two name meanings (e.g. first name meaning + surname
// meaning) without inserting extra connector words.
func CombineMeanings(m1, m2 string) string {
	m1 = strings.TrimSpace(m1)
	m2 = strings.TrimSpace(m2)
	if m1 == "" {
		return m2
	}
	if m2 == "" {
		return m1
	}
	if m1 == m2 {
		return m1
	}

	cleanM2 := m2
	if strings.HasPrefix(cleanM2, "และ") {
		cleanM2 = strings.TrimSpace(strings.TrimPrefix(cleanM2, "และ"))
	}
	if strings.HasPrefix(cleanM2, "ผู้") {
		cleanM2 = strings.TrimPrefix(cleanM2, "ผู้")
	}

	return m1 + " " + cleanM2
}

type Celebrity struct {
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

type NameSearchUseCase interface {
	SearchTopNames(ctx context.Context, req NameSearchRequest) ([]NameResultItem, error)
	SearchTopNamesWithToken(ctx context.Context, req NameSearchRequest) (NameSearchResult, error)
	RerankTopNames(ctx context.Context, req NameRerankRequest) (NameSearchResult, error)
	SuggestNames(ctx context.Context, seedName string, limit int) ([]NameResultItem, error)
	GetAllNumbersMap(ctx context.Context) (map[string]NumberMeaning, error)
	GetSumsByTHName(ctx context.Context, thname string) (int, int, string, error)
	GetSumsForText(ctx context.Context, text string) (int, int, error)
	GetSampleNames(ctx context.Context) ([]Celebrity, error)
	GetRandomMeanings(ctx context.Context, limit int) ([]string, error)
}
