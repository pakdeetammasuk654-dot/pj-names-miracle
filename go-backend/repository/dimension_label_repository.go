package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go-backend/domain"
)

type DimensionLabelRepository interface {
	GetAll(ctx context.Context) ([]domain.DimensionLabel, error)
	GetGrouped(ctx context.Context) ([]domain.DimensionGroup, error)
	GetLabelForScore(ctx context.Context, dimension string, score float64) string
	Update(ctx context.Context, id int, minScore float64, label, description string) error
	InvalidateCache()
}

type dimensionLabelRepository struct {
	db        *sql.DB
	cacheMu   sync.RWMutex
	cached    []domain.DimensionLabel
	lastFetch time.Time
}

func NewDimensionLabelRepository(db *sql.DB) DimensionLabelRepository {
	return &dimensionLabelRepository{db: db}
}

func (r *dimensionLabelRepository) GetAll(ctx context.Context) ([]domain.DimensionLabel, error) {
	r.cacheMu.RLock()
	if len(r.cached) > 0 && time.Since(r.lastFetch) < 2*time.Minute {
		defer r.cacheMu.RUnlock()
		items := make([]domain.DimensionLabel, len(r.cached))
		copy(items, r.cached)
		return items, nil
	}
	r.cacheMu.RUnlock()

	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()

	// Double check after lock
	if len(r.cached) > 0 && time.Since(r.lastFetch) < 2*time.Minute {
		items := make([]domain.DimensionLabel, len(r.cached))
		copy(items, r.cached)
		return items, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, dimension, min_score, label, COALESCE(description, ''), COALESCE(sort_order, 0), updated_at
		FROM dimension_labels
		ORDER BY dimension ASC, min_score DESC, sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query dimension_labels: %w", err)
	}
	defer rows.Close()

	var items []domain.DimensionLabel
	for rows.Next() {
		var item domain.DimensionLabel
		if err := rows.Scan(
			&item.ID,
			&item.Dimension,
			&item.MinScore,
			&item.Label,
			&item.Description,
			&item.SortOrder,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dimension_label: %w", err)
		}
		items = append(items, item)
	}

	r.cached = items
	r.lastFetch = time.Now()

	res := make([]domain.DimensionLabel, len(items))
	copy(res, items)
	return res, nil
}

func (r *dimensionLabelRepository) GetGrouped(ctx context.Context) ([]domain.DimensionGroup, error) {
	items, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	dimMeta := map[string]struct {
		Title string
		Icon  string
	}{
		"meaning":    {Title: "ความหมาย (Meaning)", Icon: "🎯"},
		"numerology": {Title: "เลขศาสตร์ (Numerology)", Icon: "⭐"},
		"sound":      {Title: "ไพเราะ (Sound / Phonetics)", Icon: "🎶"},
		"resonance":  {Title: "เอกลักษณ์ (Morphology & Identity)", Icon: "🔍"},
	}

	order := []string{"meaning", "numerology", "resonance", "sound"}
	groupMap := make(map[string][]domain.DimensionLabel)
	for _, it := range items {
		groupMap[it.Dimension] = append(groupMap[it.Dimension], it)
	}

	var groups []domain.DimensionGroup
	for _, dim := range order {
		meta := dimMeta[dim]
		groups = append(groups, domain.DimensionGroup{
			Dimension: dim,
			Title:     meta.Title,
			Icon:      meta.Icon,
			Labels:    groupMap[dim],
		})
	}
	return groups, nil
}

func (r *dimensionLabelRepository) GetLabelForScore(ctx context.Context, dimension string, score float64) string {
	items, err := r.GetAll(ctx)
	if err != nil || len(items) == 0 {
		return defaultLabelFallback(dimension, score)
	}

	var fallback string
	for _, it := range items {
		if it.Dimension == dimension {
			if fallback == "" {
				fallback = it.Label
			}
			if score >= it.MinScore {
				return it.Label
			}
		}
	}
	if fallback != "" {
		return fallback
	}
	return defaultLabelFallback(dimension, score)
}

func (r *dimensionLabelRepository) Update(ctx context.Context, id int, minScore float64, label, description string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dimension_labels
		SET min_score = $1, label = $2, description = $3, updated_at = NOW()
		WHERE id = $4
	`, minScore, label, description, id)
	if err != nil {
		return fmt.Errorf("update dimension_label: %w", err)
	}
	r.InvalidateCache()
	return nil
}

func (r *dimensionLabelRepository) InvalidateCache() {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	r.cached = nil
	r.lastFetch = time.Time{}
}

func defaultLabelFallback(dimension string, score float64) string {
	switch dimension {
	case "meaning":
		if score >= 90 {
			return "โดดเด่นระดับสูง"
		} else if score >= 80 {
			return "เด่นมาก มงคลลึกซึ้ง"
		} else if score >= 70 {
			return "ความหมายดี มงคล"
		}
		return "ความหมายทั่วไป"
	case "numerology":
		if score >= 90 {
			return "คู่มงคลชั้นสูง มหาบารมี"
		} else if score >= 80 {
			return "คู่มงคลพลังดีเยี่ยม"
		} else if score >= 70 {
			return "คู่มงคลเกณฑ์ดี"
		}
		return "คู่เลขพลังปานกลาง"
	case "sound":
		if score >= 90 {
			return "จังหวะเสียงไพเราะก้องกังวาน"
		} else if score >= 80 {
			return "เสียงสง่างาม มีน้ำหนัก"
		} else if score >= 70 {
			return "ไพเราะ คล่องตัว"
		}
		return "เสียงยาว ลุ่มลึก"
	case "resonance":
		if score >= 88 {
			return "สมดุลรอบด้าน เอกลักษณ์สูง"
		} else if score >= 75 {
			return "กลมกลืน ทรงเอกลักษณ์"
		}
		return "เอกลักษณ์เฉพาะตัว"
	}
	return "มาตรฐาน"
}
