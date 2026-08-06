package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"go-backend/domain"
)

var (
	vectorCache   = make(map[string]string)
	vectorCacheMu sync.RWMutex
)

const maxVectorCacheEntries = 256

type NameRepository interface {
	GetGoodNumbersMap(ctx context.Context) (map[int]domain.NumberMeaning, error)
	GetAllNumbersMap(ctx context.Context) (map[int]domain.NumberMeaning, error)
	GetSurnameSums(ctx context.Context, surname string) (satSum int, shaSum int, err error)
	GetSumsByTHName(ctx context.Context, thname string) (satSum int, shaSum int, meaning string, err error)
	SearchCandidateNames(ctx context.Context, targetSatSums []int, targetShaSums []int, dayOfBirth string, wishQuery string, limit int) ([]domain.NameResultItem, error)
	SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error)
	SearchSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, dayOfBirth string, limit int) ([]domain.NameResultItem, error)
	GetSampleNames(ctx context.Context) ([]domain.Celebrity, error)
}

type nameRepository struct {
	db *sql.DB
}

func NewNameRepository(db *sql.DB) NameRepository {
	return &nameRepository{db: db}
}

// GetGoodNumbersMap fetches numbers table and returns map of integer pairnumber -> NumberMeaning for D10, D8, D5
func (r *nameRepository) GetGoodNumbersMap(ctx context.Context) (map[int]domain.NumberMeaning, error) {
	query := `SELECT pairnumber, pairtype, pairpoint, COALESCE(miracledesc, ''), COALESCE(miracledetail, '') FROM numbers WHERE pairtype LIKE 'D%'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch good numbers dictionary: %w", err)
	}
	defer rows.Close()

	result := make(map[int]domain.NumberMeaning)
	for rows.Next() {
		var n domain.NumberMeaning
		if err := rows.Scan(&n.PairNumber, &n.PairType, &n.PairPoint, &n.MiracleDesc, &n.MiracleDetail); err != nil {
			return nil, err
		}
		var val int
		fmt.Sscanf(n.PairNumber, "%d", &val)
		result[val] = n
	}
	return result, nil
}

// GetAllNumbersMap fetches ALL 110 pair numbers (both D-type and R-type) from numbers table
func (r *nameRepository) GetAllNumbersMap(ctx context.Context) (map[int]domain.NumberMeaning, error) {
	query := `SELECT pairnumber, pairtype, pairpoint, COALESCE(miracledesc, ''), COALESCE(miracledetail, '') FROM numbers`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all numbers dictionary: %w", err)
	}
	defer rows.Close()

	result := make(map[int]domain.NumberMeaning)
	for rows.Next() {
		var n domain.NumberMeaning
		if err := rows.Scan(&n.PairNumber, &n.PairType, &n.PairPoint, &n.MiracleDesc, &n.MiracleDetail); err != nil {
			return nil, err
		}
		var val int
		fmt.Sscanf(n.PairNumber, "%d", &val)
		result[val] = n
	}
	return result, nil
}

// GetSurnameSums decodes ANY Thai surname dynamically using sat_nums and sha_nums tables
func (r *nameRepository) GetSurnameSums(ctx context.Context, surname string) (int, int, error) {
	if surname == "" {
		return 0, 0, nil
	}

	var satSum, shaSum int
	directQuery := `SELECT sat_sum, sha_sum FROM names_miracle WHERE thname = $1 LIMIT 1`
	err := r.db.QueryRowContext(ctx, directQuery, surname).Scan(&satSum, &shaSum)
	if err == nil {
		return satSum, shaSum, nil
	}

	decodeQuery := `
		SELECT 
			COALESCE(SUM(s.sat_value), 0) as sat_total, 
			COALESCE(SUM(sh.sha_value), 0) as sha_total 
		FROM regexp_split_to_table($1, '') as c 
		LEFT JOIN sat_nums s ON s.char_key = c 
		LEFT JOIN sha_nums sh ON sh.char_key = c
	`
	err = r.db.QueryRowContext(ctx, decodeQuery, surname).Scan(&satSum, &shaSum)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode surname numerology: %w", err)
	}

	return satSum, shaSum, nil
}

// GetSumsByTHName gets the pre-calculated SatSum and ShaSum of a specific name
func (r *nameRepository) GetSumsByTHName(ctx context.Context, thname string) (int, int, string, error) {
	var satSum, shaSum int
	var meaning string
	err := r.db.QueryRowContext(ctx, "SELECT sat_sum, sha_sum, COALESCE(meaning, '') FROM names_miracle WHERE thname = $1 LIMIT 1", thname).Scan(&satSum, &shaSum, &meaning)
	if err != nil {
		return 0, 0, "", err
	}
	return satSum, shaSum, meaning, nil
}

// SearchCandidateNames fetches candidates from Postgres names_miracle table
func (r *nameRepository) SearchCandidateNames(ctx context.Context, targetSatSums []int, targetShaSums []int, dayOfBirth string, wishQuery string, limit int) ([]domain.NameResultItem, error) {
	if len(targetSatSums) == 0 && len(targetShaSums) == 0 {
		return []domain.NameResultItem{}, nil
	}

	validDays := map[string]string{
		"sunday":     "k_sunday",
		"monday":     "k_monday",
		"tuesday":    "k_tuesday",
		"wednesday1": "k_wednesday1",
		"wednesday2": "k_wednesday2",
		"thursday":   "k_thursday",
		"friday":     "k_friday",
		"saturday":   "k_saturday",
	}

	whereClause := "sat_sum = ANY($1) AND sha_sum = ANY($2)"
	dayClean := strings.ToLower(strings.TrimSpace(dayOfBirth))

	if col, ok := validDays[dayClean]; ok {
		whereClause += fmt.Sprintf(" AND %s = false", col)
	}

	wishClean := strings.TrimSpace(wishQuery)
	var vecStr string
	if wishClean != "" {
		v, err := generateOpenAIVector(ctx, wishClean)
		if err == nil && v != "" {
			vecStr = v
		}
	}

	tokens := extractThaiTokens(wishClean)
	var kwClause string
	if len(tokens) > 0 {
		var sub []string
		for _, tok := range tokens {
			sub = append(sub, fmt.Sprintf("meaning ILIKE '%%%s%%'", strings.ReplaceAll(tok, "'", "''")))
		}
		kwClause = fmt.Sprintf(" AND (%s)", strings.Join(sub, " OR "))
	}

	var query string
	var rows *sql.Rows
	var err error

	if vecStr != "" {
		query = fmt.Sprintf(`
			SELECT 
				name_id, thname, COALESCE(meaning, ''), sat_sum, sha_sum,
				(meaning_vector <=> $4::vector) as distance
			FROM names_miracle 
			WHERE %s AND meaning_vector IS NOT NULL
			ORDER BY (meaning_vector <=> $4::vector) ASC
			LIMIT $3
		`, whereClause)
		rows, err = r.db.QueryContext(ctx, query, pq.Array(targetSatSums), pq.Array(targetShaSums), 1000, vecStr)
	} else if kwClause != "" {
		query = fmt.Sprintf(`
			SELECT name_id, thname, COALESCE(meaning, ''), sat_sum, sha_sum, 0.0 as distance
			FROM names_miracle 
			WHERE %s %s
			LIMIT $3
		`, whereClause, kwClause)
		rows, err = r.db.QueryContext(ctx, query, pq.Array(targetSatSums), pq.Array(targetShaSums), 1000)
	} else {
		query = fmt.Sprintf(`
			SELECT name_id, thname, COALESCE(meaning, ''), sat_sum, sha_sum, 0.0 as distance
			FROM names_miracle 
			WHERE %s
			LIMIT $3
		`, whereClause)
		rows, err = r.db.QueryContext(ctx, query, pq.Array(targetSatSums), pq.Array(targetShaSums), 1000)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query candidate names: %w", err)
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum, &item.VectorDistance); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// SearchSemanticWishNames converts user wish to 1536-dim OpenAI Vector & queries pgvector HNSW Index
func (r *nameRepository) SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error) {
	validDays := map[string]string{
		"sunday":     "k_sunday",
		"monday":     "k_monday",
		"tuesday":    "k_tuesday",
		"wednesday1": "k_wednesday1",
		"wednesday2": "k_wednesday2",
		"thursday":   "k_thursday",
		"friday":     "k_friday",
		"saturday":   "k_saturday",
	}

	dayClean := strings.ToLower(strings.TrimSpace(dayOfBirth))
	dayClause := ""
	if col, ok := validDays[dayClean]; ok {
		dayClause = fmt.Sprintf(" AND %s = false", col)
	}

	if limit <= 0 {
		limit = 100
	}

	// 1. Real-time OpenAI text-embedding-3-small vector generation
	vecStr, err := generateOpenAIVector(ctx, wishQuery)
	if err != nil {
	} else {
	}

	if err == nil && vecStr != "" {
		vectorQuery := fmt.Sprintf(`
			SELECT 
				name_id, 
				thname, 
				COALESCE(meaning, ''), 
				sat_sum, 
				sha_sum,
				(meaning_vector <=> $1::vector) as distance
			FROM names_miracle 
			WHERE meaning_vector IS NOT NULL 
			AND sat_sum = ANY($3) AND sha_sum = ANY($4)
			%s
			ORDER BY meaning_vector <=> $1::vector ASC
			LIMIT $2
		`, dayClause)

		tx, err := r.db.BeginTx(ctx, nil)
		if err == nil {
			defer tx.Rollback()
			// Force Exact Search for vector KNN when pre-filtering with highly selective WHERE clauses
			// This prevents the HNSW index from terminating early and returning too few candidates.
			tx.ExecContext(ctx, "SET LOCAL enable_indexscan = off;")

			rows, err := tx.QueryContext(ctx, vectorQuery, vecStr, 1000, pq.Array(targetSatSums), pq.Array(targetShaSums))
			if err == nil {
				defer rows.Close()
				var results []domain.NameResultItem
				for rows.Next() {
					var item domain.NameResultItem
					if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum, &item.VectorDistance); err == nil {
						results = append(results, item)
					}
				}
				tx.Commit()
				if len(results) > 0 {
					return results, nil
				}
			}
		}
	}

	// 2. Fallback to Seed Vector lookup if OpenAI API fails or is offline
	tokens := extractThaiTokens(strings.TrimSpace(wishQuery))
	firstKw := "ร่ำรวย"
	if len(tokens) > 0 {
		firstKw = tokens[0]
	}

	var seedVectorStr string
	seedQuery := fmt.Sprintf(`
		SELECT meaning_vector::text 
		FROM names_miracle 
		WHERE meaning ILIKE '%%%s%%' AND meaning_vector IS NOT NULL 
		LIMIT 1
	`, strings.ReplaceAll(firstKw, "'", "''"))

	err = r.db.QueryRowContext(ctx, seedQuery).Scan(&seedVectorStr)
	if err == nil && seedVectorStr != "" {
		vectorQuery := fmt.Sprintf(`
			SELECT 
				name_id, 
				thname, 
				COALESCE(meaning, ''), 
				sat_sum, 
				sha_sum,
				(meaning_vector <=> $1::vector) as distance
			FROM names_miracle 
			WHERE meaning_vector IS NOT NULL %s
			ORDER BY meaning_vector <=> $1::vector ASC
			LIMIT $2
		`, dayClause)

		rows, err := r.db.QueryContext(ctx, vectorQuery, seedVectorStr, limit*2)
		if err == nil {
			defer rows.Close()
			var results []domain.NameResultItem
			for rows.Next() {
				var item domain.NameResultItem
				if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum, &item.VectorDistance); err == nil {
					results = append(results, item)
				}
			}
			if len(results) > 0 {
				return results, nil
			}
		}
	}

	return r.searchSemanticTextFallback(ctx, tokens, dayClause, limit)
}

// generateOpenAIVector calls OpenAI text-embedding-3-small API with In-Memory Caching (0ms latency for repeated/preset queries)
func generateOpenAIVector(ctx context.Context, text string) (string, error) {
	cleanText := strings.TrimSpace(text)

	// Check In-Memory Cache first
	vectorCacheMu.RLock()
	if cachedVec, found := vectorCache[cleanText]; found {
		vectorCacheMu.RUnlock()
		return cachedVec, nil
	}
	vectorCacheMu.RUnlock()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"input": "ความหมาย: " + text,
		"model": "text-embedding-3-small",
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api error status: %d", resp.StatusCode)
	}

	var res struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil || len(res.Data) == 0 {
		return "", fmt.Errorf("failed to decode openai response")
	}

	vecStrBytes, _ := json.Marshal(res.Data[0].Embedding)
	vecResult := string(vecStrBytes)

	// Save to In-Memory Cache
	vectorCacheMu.Lock()
	if len(vectorCache) >= maxVectorCacheEntries {
		// Preset wishes are requested repeatedly. Bounding this cache prevents an
		// unbounded stream of custom wishes from growing the API process forever.
		vectorCache = make(map[string]string)
	}
	vectorCache[cleanText] = vecResult
	vectorCacheMu.Unlock()

	return vecResult, nil
}

func (r *nameRepository) searchSemanticTextFallback(ctx context.Context, tokens []string, dayClause string, limit int) ([]domain.NameResultItem, error) {
	baseWhere := "meaning IS NOT NULL AND meaning != ''" + dayClause
	var subClauses []string
	for _, tok := range tokens {
		subClauses = append(subClauses, fmt.Sprintf("meaning ILIKE '%%%s%%'", strings.ReplaceAll(tok, "'", "''")))
	}

	whereClause := baseWhere
	if len(subClauses) > 0 {
		whereClause += fmt.Sprintf(" AND (%s)", strings.Join(subClauses, " OR "))
	}

	query := fmt.Sprintf(`
		SELECT name_id, thname, COALESCE(meaning, ''), sat_sum, sha_sum 
		FROM names_miracle 
		WHERE %s
		LIMIT $1
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum); err == nil {
			results = append(results, item)
		}
	}
	return results, nil
}

// extractThaiTokens splits sentences into meaningful Thai keywords
func extractThaiTokens(text string) []string {
	commonDict := []string{
		"ร่ำรวย", "เศรษฐี", "มั่งคั่ง", "ทรัพย์", "โชค", "ลาภ",
		"ปัญญา", "เฉลียวฉลาด", "ความรู้", "เก่ง", "ฉลาด",
		"ผู้นำ", "อำนาจ", "บารมี", "ยศ", "ศักดิ์", "ตำแหน่ง",
		"เมตตา", "เสน่ห์", "สิริมงคล", "ความสุข", "เจริญ", "รุ่งเรือง",
		"แข็งแรง", "อายุยืน", "บริบูรณ์", "สมบูรณ์", "ประเสริฐ",
	}

	var found []string
	for _, word := range commonDict {
		if strings.Contains(text, word) {
			found = append(found, word)
		}
	}

	if len(found) == 0 {
		fields := strings.Fields(text)
		if len(fields) > 0 {
			return fields
		}
		if len(text) > 0 {
			return []string{text}
		}
	}
	return found
}

// SearchSimilarNames uses PostgreSQL pg_trgm similarity to rank names that are textually similar to the seedName
func (r *nameRepository) SearchSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, dayOfBirth string, limit int) ([]domain.NameResultItem, error) {
	validDays := map[string]string{
		"sunday":     "k_sunday",
		"monday":     "k_monday",
		"tuesday":    "k_tuesday",
		"wednesday1": "k_wednesday1",
		"wednesday2": "k_wednesday2",
		"thursday":   "k_thursday",
		"friday":     "k_friday",
		"saturday":   "k_saturday",
	}

	dayClean := strings.ToLower(strings.TrimSpace(dayOfBirth))
	dayClause := ""
	if col, ok := validDays[dayClean]; ok {
		dayClause = fmt.Sprintf(" AND %s = false", col)
	}

	if limit <= 0 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT 
			name_id, 
			thname, 
			COALESCE(meaning, ''), 
			sat_sum, 
			sha_sum,
			(1.0 - similarity(thname, $1)) as distance
		FROM names_miracle 
		WHERE sat_sum = ANY($2) AND sha_sum = ANY($3)
		%s
		ORDER BY similarity(thname, $1) DESC
		LIMIT $4
	`, dayClause)

	rows, err := r.db.QueryContext(ctx, query, seedName, pq.Array(targetSatSums), pq.Array(targetShaSums), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum, &item.VectorDistance); err == nil {
			results = append(results, item)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *nameRepository) GetSampleNames(ctx context.Context) ([]domain.Celebrity, error) {
	query := `SELECT name, avatar_url FROM naming_examples WHERE is_celebrity = true ORDER BY sort_order ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []domain.Celebrity
	for rows.Next() {
		var name, avatarUrl string
		if err := rows.Scan(&name, &avatarUrl); err != nil {
			return nil, err
		}

		// Extract filename from avatarUrl (e.g. /uploads/avatars/avatar_123.jpg -> avatar_123.jpg)
		parts := strings.Split(avatarUrl, "/")
		filename := parts[len(parts)-1]

		names = append(names, domain.Celebrity{
			Name:      name,
			AvatarUrl: filename,
		})
	}
	return names, nil
}
