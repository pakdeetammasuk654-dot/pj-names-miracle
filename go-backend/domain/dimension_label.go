package domain

import "time"

type DimensionLabel struct {
	ID          int       `json:"id"`
	Dimension   string    `json:"dimension"`   // meaning, numerology, sound, resonance
	MinScore    float64   `json:"min_score"`   // e.g. 90.0, 80.0, 70.0, 0.0
	Label       string    `json:"label"`       // e.g. "โดดเด่นระดับสูง"
	Description string    `json:"description"` // explanation
	SortOrder   int       `json:"sort_order"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DimensionGroup struct {
	Dimension string           `json:"dimension"`
	Title     string           `json:"title"`
	Icon      string           `json:"icon"`
	Labels    []DimensionLabel `json:"labels"`
}
