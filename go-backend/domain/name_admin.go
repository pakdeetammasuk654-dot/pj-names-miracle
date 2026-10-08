package domain

import (
	"time"
)

// AdminNameItem represents a name record in the admin management interface.
type AdminNameItem struct {
	NameID             int64     `json:"name_id"`
	THName             string    `json:"thname"`
	Meaning            string    `json:"meaning"`
	Gender             string    `json:"gender"`
	SatSum             int       `json:"sat_sum"`
	ShaSum             int       `json:"sha_sum"`
	SatPairs           []string  `json:"sat_pairs"`
	ShaPairs           []string  `json:"sha_pairs"`
	KSunday            bool      `json:"k_sunday"`
	KMonday            bool      `json:"k_monday"`
	KTuesday           bool      `json:"k_tuesday"`
	KWednesday1        bool      `json:"k_wednesday1"` // พุธกลางวัน
	KWednesday2        bool      `json:"k_wednesday2"` // พุธกลางคืน
	KThursday          bool      `json:"k_thursday"`
	KFriday            bool      `json:"k_friday"`
	KSaturday          bool      `json:"k_saturday"`
	EuphonyScore       int       `json:"euphony_score"`
	PronunciationEase  int       `json:"pronunciation_ease"`
	RhythmScore        int       `json:"rhythm_score"`
	PhoneticVersion    string    `json:"phonetic_version"`
	HasVector          bool      `json:"has_vector"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
}

// DayKalakiniStatus describes whether a name is auspicious for a specific day of birth.
type DayKalakiniStatus struct {
	Field    string `json:"field"`     // e.g. "k_sunday"
	DayTH    string `json:"day_th"`    // e.g. "วันอาทิตย์"
	IsKakis  bool   `json:"is_kakis"`  // true if contains Kalakini letters (incompatible)
}

// NameAnalysisRequest is the payload sent from admin to analyze a candidate name.
type NameAnalysisRequest struct {
	THName      string `json:"thname"`
	AutoMeaning bool   `json:"auto_meaning"`
	Gender      string `json:"gender"`
}

// NameAnalysisResult contains calculated numerology, Kalakini days, and AI suggestions.
type NameAnalysisResult struct {
	THName            string              `json:"thname"`
	SatSum            int                 `json:"sat_sum"`
	ShaSum            int                 `json:"sha_sum"`
	SatPairs          []string            `json:"sat_pairs"`
	ShaPairs          []string            `json:"sha_pairs"`
	Days              []DayKalakiniStatus `json:"days"`
	Meaning           string              `json:"meaning"`
	Gender            string              `json:"gender"`
	EuphonyScore      int                 `json:"euphony_score"`
	PronunciationEase int                 `json:"pronunciation_ease"`
	RhythmScore       int                 `json:"rhythm_score"`
	AlreadyExists     bool                `json:"already_exists"`
	ExistingID        int64               `json:"existing_id,omitempty"`
}

// CreateAdminNameRequest is the form submission to persist a complete name.
type CreateAdminNameRequest struct {
	THName            string `json:"thname"`
	Meaning           string `json:"meaning"`
	Gender            string `json:"gender"`
	SatSum            int    `json:"sat_sum"`
	ShaSum            int    `json:"sha_sum"`
	EuphonyScore      int    `json:"euphony_score"`
	PronunciationEase int    `json:"pronunciation_ease"`
	RhythmScore       int    `json:"rhythm_score"`
}
