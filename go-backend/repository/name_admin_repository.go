package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lib/pq"
	"go-backend/domain"
)

type NameAdminRepository interface {
	GetDictionary(ctx context.Context) (map[rune]int, map[rune]int, error)
	GetKalakiniRules(ctx context.Context) (map[string]map[rune]bool, error)
	ExistsTHName(ctx context.Context, thname string) (bool, int64, error)
	InsertCompleteName(ctx context.Context, item *domain.AdminNameItem, vectorJSON string) (int64, error)
	GetRecentNames(ctx context.Context, search string, limit int, offset int) ([]domain.AdminNameItem, int, error)
	GetTotalNamesCount(ctx context.Context) (int, error)
}

type nameAdminRepository struct {
	db        *sql.DB
	dictOnce  sync.Once
	satDict   map[rune]int
	shaDict   map[rune]int
	dictErr   error
	kakisOnce sync.Once
	kakisMap  map[string]map[rune]bool
	kakisErr  error
}

func NewNameAdminRepository(db *sql.DB) NameAdminRepository {
	return &nameAdminRepository{db: db}
}

func (r *nameAdminRepository) GetDictionary(ctx context.Context) (map[rune]int, map[rune]int, error) {
	r.dictOnce.Do(func() {
		rows, err := r.db.QueryContext(ctx, `
			SELECT COALESCE(s.char_key, sh.char_key), COALESCE(s.sat_value, 0), COALESCE(sh.sha_value, 0)
			FROM sat_nums AS s FULL OUTER JOIN sha_nums AS sh ON sh.char_key = s.char_key`)
		if err != nil {
			r.dictErr = err
			return
		}
		defer rows.Close()

		satMap := make(map[rune]int)
		shaMap := make(map[rune]int)
		for rows.Next() {
			var key string
			var sat, sha int
			if err := rows.Scan(&key, &sat, &sha); err != nil {
				r.dictErr = err
				return
			}
			if utf8.RuneCountInString(key) == 1 {
				rn := []rune(key)[0]
				satMap[rn] = sat
				shaMap[rn] = sha
			}
		}
		if err := rows.Err(); err != nil {
			r.dictErr = err
			return
		}
		r.satDict = satMap
		r.shaDict = shaMap
	})
	return r.satDict, r.shaDict, r.dictErr
}

func (r *nameAdminRepository) GetKalakiniRules(ctx context.Context) (map[string]map[rune]bool, error) {
	r.kakisOnce.Do(func() {
		days := map[string][]string{
			"k_sunday":       {"Sunday", "อาทิตย์"},
			"k_monday":       {"Monday", "จันทร์"},
			"k_tuesday":      {"Tuesday", "อังคาร"},
			"k_wednesday1":   {"Wednesday1", "พุธกลางวัน"},
			"k_wednesday2":   {"Wednesday2", "พุธกลางคืน"},
			"k_thursday":     {"Thursday", "พฤหัสบดี"},
			"k_friday":       {"Friday", "ศุกร์"},
			"k_saturday":     {"Saturday", "เสาร์"},
		}
		result := make(map[string]map[rune]bool, len(days))
		for field, names := range days {
			rows, err := r.db.QueryContext(ctx, `SELECT kakis FROM kakis_day WHERE day = $1 OR day_th = $2`, names[0], names[1])
			if err != nil {
				r.kakisErr = err
				return
			}
			chars := make(map[rune]bool)
			for rows.Next() {
				var text string
				if err := rows.Scan(&text); err != nil {
					rows.Close()
					r.kakisErr = err
					return
				}
				for _, ch := range strings.TrimSpace(text) {
					chars[ch] = true
				}
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				r.kakisErr = err
				return
			}
			result[field] = chars
		}
		r.kakisMap = result
	})
	return r.kakisMap, r.kakisErr
}

func (r *nameAdminRepository) ExistsTHName(ctx context.Context, thname string) (bool, int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT name_id FROM names_miracle WHERE thname = $1 LIMIT 1`, thname).Scan(&id)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

func (r *nameAdminRepository) InsertCompleteName(ctx context.Context, item *domain.AdminNameItem, vectorJSON string) (int64, error) {
	query := `
		INSERT INTO names_miracle (
			thname, meaning, gender, sat_sum, sha_sum,
			satnum, shanum, t_sat, t_sha,
			k_sunday, k_monday, k_tuesday, k_wednesday1, k_wednesday2, k_thursday, k_friday, k_saturday,
			euphony_score, pronunciation_ease, rhythm_score, phonetic_version,
			meaning_vector
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21,
			CASE WHEN $22 <> '' THEN $22::vector ELSE NULL END
		) RETURNING name_id`

	var nameID int64
	err := r.db.QueryRowContext(ctx, query,
		item.THName, item.Meaning, item.Gender, item.SatSum, item.ShaSum,
		pq.Array(item.SatPairs), pq.Array(item.ShaPairs), pq.Array(item.SatPairs), pq.Array(item.ShaPairs),
		item.KSunday, item.KMonday, item.KTuesday, item.KWednesday1, item.KWednesday2, item.KThursday, item.KFriday, item.KSaturday,
		item.EuphonyScore, item.PronunciationEase, item.RhythmScore, item.PhoneticVersion,
		vectorJSON,
	).Scan(&nameID)

	return nameID, err
}

func (r *nameAdminRepository) GetRecentNames(ctx context.Context, search string, limit int, offset int) ([]domain.AdminNameItem, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	whereClause := ""
	var args []any
	countArgs := []any{}
	if strings.TrimSpace(search) != "" {
		whereClause = "WHERE thname LIKE '%' || $1 || '%' OR meaning LIKE '%' || $1 || '%'"
		args = append(args, strings.TrimSpace(search))
		countArgs = append(countArgs, strings.TrimSpace(search))
	}

	var total int
	countQuery := fmt.Sprintf("SELECT count(*) FROM names_miracle %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT name_id, thname, COALESCE(meaning, ''), COALESCE(gender, 'ค'), sat_sum, sha_sum,
			k_sunday, k_monday, k_tuesday, k_wednesday1, k_wednesday2, k_thursday, k_friday, k_saturday,
			COALESCE(euphony_score, 0), COALESCE(pronunciation_ease, 0), COALESCE(rhythm_score, 0),
			meaning_vector IS NOT NULL
		FROM names_miracle
		%s
		ORDER BY name_id DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.AdminNameItem
	for rows.Next() {
		var item domain.AdminNameItem
		if err := rows.Scan(
			&item.NameID, &item.THName, &item.Meaning, &item.Gender, &item.SatSum, &item.ShaSum,
			&item.KSunday, &item.KMonday, &item.KTuesday, &item.KWednesday1, &item.KWednesday2, &item.KThursday, &item.KFriday, &item.KSaturday,
			&item.EuphonyScore, &item.PronunciationEase, &item.RhythmScore,
			&item.HasVector,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

func (r *nameAdminRepository) GetTotalNamesCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM names_miracle`).Scan(&count)
	return count, err
}
