package domain

import "context"

type SearchMode string

const (
	ModeCombined   SearchMode = "combined"   // ชื่อ + นามสกุล
	ModeStandalone SearchMode = "standalone" // เฉพาะชื่อ
	ModeSemantic   SearchMode = "semantic"   // ค้นหาด้วย AI Semantic Wish Search
)

type NameSearchRequest struct {
	Mode           SearchMode `json:"mode"`                      // "combined", "standalone", "semantic"
	Surname        string     `json:"surname,omitempty"`         // นามสกุล (ถ้าเป็น combined)
	DayOfBirth     string     `json:"day_of_birth,omitempty"`    // วันเกิด
	WishQuery      string     `json:"wish_query,omitempty"`      // คำอธิษฐาน/ความหมายธรรมชาติสำหรับ AI Semantic Search
	Limit          int        `json:"limit,omitempty"`           // Default 100
	SemanticWeight *float64   `json:"semantic_weight,omitempty"` // Weight for semantic vs score (0.0 = 100% Score, 1.0 = 100% Semantic)
	IsNameIntent   bool       `json:"is_name_intent,omitempty"`  // If true, WishQuery contains a seed name for Trigram textual similarity rather than semantic meaning
}

type NameResultItem struct {
	NameID            int     `json:"name_id"`
	THName            string  `json:"thname"`
	Meaning           string  `json:"meaning,omitempty"`         // ความหมายของชื่อ
	VectorDistance    float64 `json:"vector_distance,omitempty"` // ระยะทาง Vector Similarity
	SatSum            int     `json:"sat_sum"`
	ShaSum            int     `json:"sha_sum"`
	TotalSatSum       int     `json:"total_sat_sum,omitempty"`
	TotalShaSum       int     `json:"total_sha_sum,omitempty"`
	SatPairsDisplay   string  `json:"sat_pairs_display"` // ตัวเลขคู่ 00 - 99
	ShaPairsDisplay   string  `json:"sha_pairs_display"` // ตัวเลขคู่ 00 - 99
	SatPairType       string  `json:"sat_pairtype"`      // D10, D8, D5
	ShaPairType       string  `json:"sha_pairtype"`      // D10, D8, D5
	SatMiracleDesc    string  `json:"sat_miracledesc"`
	ShaMiracleDesc    string  `json:"sha_miracledesc"`
	IsPerfectCombined bool    `json:"is_perfect_combined"` // true = ผลรวม (ชื่อ + นามสกุล) ได้เกรด D10/D8/D5 สมบูรณ์แบบ
	TotalScore        int     `json:"total_score"`         // คะแนนรวมสำหรับ Rank
	FinalScore        float64 `json:"final_score"`         // คะแนนรวมแบบ Hybrid ถ่วงน้ำหนัก
	Rank              int     `json:"rank"`
}

type NameSearchPayload struct {
	Request       NameSearchRequest
	SurnameSatSum int
	SurnameShaSum int
	GoodNumbers   map[int]NumberMeaning // map[pairnumber_int]NumberMeaning
	Candidates    []NameResultItem
	FinalResults  []NameResultItem
}

type Celebrity struct {
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

type NameSearchUseCase interface {
	SearchTopNames(ctx context.Context, req NameSearchRequest) ([]NameResultItem, error)
	GetAllNumbersMap(ctx context.Context) (map[int]NumberMeaning, error)
	GetSumsByTHName(ctx context.Context, thname string) (int, int, string, error)
	GetSumsForText(ctx context.Context, text string) (int, int, error)
	GetSampleNames(ctx context.Context) ([]Celebrity, error)
}
