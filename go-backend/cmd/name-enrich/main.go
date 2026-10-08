// Command name-enrich enriches quarantined rows from name_import_staging.
// It is intentionally unable to write directly to public.names_miracle.
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
	"go-backend/config"
)

const (
	meaningPromptVersion = "meaning_gender_th_v2"
	defaultMeaningModel  = "gpt-4o-mini"
	phoneticVersion      = "typhoon_v2_5_promptA"
	defaultTyphoonModel  = "typhoon-v2.1-12b-instruct"
)

type stagedName struct {
	ID   int64
	Name string
}

type charValues struct {
	sat, sha       int
	hasSat, hasSha bool
}

type enrichmentResult struct {
	Name    string `json:"name"`
	Meaning string `json:"meaning"`
	Gender  string `json:"gender"`
	Review  bool   `json:"review_required"`
}

type phoneticResult struct {
	Name  string `json:"name"`
	Score struct {
		PronunciationEase int `json:"pronunciation_ease"`
		Euphony           int `json:"euphony"`
		Rhythm            int `json:"rhythm"`
	} `json:"score"`
	Review bool `json:"review_required"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	mode := flag.String("mode", "deterministic", "deterministic, meaning, meaning-review, phonetic, embedding, promote, or run")
	batchLabel := flag.String("batch", "", "staging batch label")
	limit := flag.Int("limit", 20, "maximum staged names to process")
	apply := flag.Bool("apply", false, "required for database writes")
	flag.Parse()

	if strings.TrimSpace(*batchLabel) == "" {
		fatal(errors.New("-batch is required"))
	}
	if *limit < 1 || *limit > 250 {
		fatal(errors.New("-limit must be between 1 and 250"))
	}
	if !*apply {
		fatal(errors.New("refusing to write: rerun with -apply after reviewing the selected mode"))
	}

	db, err := config.InitDB()
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	batchID, err := findBatchID(db, strings.TrimSpace(*batchLabel))
	if err != nil {
		fatal(err)
	}

	switch *mode {
	case "deterministic":
		processed, review, err := enrichDeterministic(db, batchID, *limit)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Deterministic enrichment completed: processed=%d review_required=%d\n", processed, review)
	case "meaning":
		processed, review, err := enrichMeaning(context.Background(), db, batchID, *limit, false)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Meaning enrichment completed: processed=%d review_required=%d\n", processed, review)
	case "meaning-review":
		processed, review, err := enrichMeaning(context.Background(), db, batchID, *limit, true)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Meaning review enrichment completed: processed=%d review_required=%d\n", processed, review)
	case "phonetic":
		processed, review, err := enrichPhonetic(context.Background(), db, batchID, *limit)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Phonetic enrichment completed: processed=%d review_required=%d\n", processed, review)
	case "embedding":
		processed, err := enrichEmbeddings(context.Background(), db, batchID, *limit)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Embedding enrichment completed: processed=%d\n", processed)
	case "promote":
		processed, err := promoteCompleteRows(db, batchID, *limit)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("Promotion completed: inserted=%d\n", processed)
	case "run":
		if err := runEnrichment(ctxOrBackground(), db, batchID, *limit); err != nil {
			fatal(err)
		}
		fmt.Println("Enrichment completed. Promotion remains a separate reviewed step.")
	default:
		fatal(fmt.Errorf("unsupported -mode %q", *mode))
	}
}

func ctxOrBackground() context.Context { return context.Background() }

func runEnrichment(ctx context.Context, db *sql.DB, batchID int64, batchSize int) error {
	for {
		processed, review, err := enrichDeterministic(db, batchID, batchSize)
		if err != nil {
			return err
		}
		if processed+review == 0 {
			break
		}
		fmt.Printf("[run] deterministic processed=%d review_required=%d\n", processed, review)
	}
	for {
		processed, review, err := enrichMeaning(ctx, db, batchID, batchSize, false)
		if err != nil {
			return err
		}
		if processed+review == 0 {
			break
		}
		fmt.Printf("[run] meaning processed=%d review_required=%d\n", processed, review)
	}
	for {
		processed, review, err := enrichMeaning(ctx, db, batchID, batchSize, true)
		if err != nil {
			return err
		}
		if processed+review == 0 {
			break
		}
		fmt.Printf("[run] meaning-review processed=%d review_required=%d\n", processed, review)
	}
	for {
		// A phonetic result contains three scores per name. Keep this request
		// deliberately smaller than the other enrichment phases so the model
		// cannot truncate its JSON response mid-batch.
		phoneticBatchSize := batchSize
		if phoneticBatchSize > 12 {
			phoneticBatchSize = 12
		}
		processed, review, err := enrichPhonetic(ctx, db, batchID, phoneticBatchSize)
		if err != nil {
			return err
		}
		if processed+review == 0 {
			break
		}
		fmt.Printf("[run] phonetic processed=%d review_required=%d\n", processed, review)
	}
	for {
		processed, err := enrichEmbeddings(ctx, db, batchID, batchSize)
		if err != nil {
			return err
		}
		if processed == 0 {
			break
		}
		fmt.Printf("[run] embedding processed=%d\n", processed)
	}
	return nil
}

func findBatchID(db *sql.DB, label string) (int64, error) {
	var id int64
	err := db.QueryRow(`SELECT batch_id FROM data_governance.name_import_batches WHERE batch_label = $1`, label).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("unknown staging batch %q", label)
	}
	return id, err
}

func enrichDeterministic(db *sql.DB, batchID int64, limit int) (processed, review int, err error) {
	dictionary, err := loadDictionary(db)
	if err != nil {
		return 0, 0, err
	}
	kalakini, err := loadKalakini(db)
	if err != nil {
		return 0, 0, err
	}
	rows, err := pendingRows(db, batchID, "sat_sum IS NULL OR sha_sum IS NULL", limit)
	if err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		sat, sha, unknown := calculateSums(row.Name, dictionary)
		if len(unknown) > 0 {
			if err := markReview(db, row.ID, "unmapped_characters:"+strings.Join(unknown, "")); err != nil {
				return processed, review, err
			}
			review++
			continue
		}
		flags := calculateKalakini(row.Name, kalakini)
		_, err := db.Exec(`
			UPDATE data_governance.name_import_staging
			SET sat_sum=$1, sha_sum=$2,
				k_sunday=$3, k_monday=$4, k_tuesday=$5, k_wednesday1=$6,
				k_wednesday2=$7, k_thursday=$8, k_friday=$9, k_saturday=$10,
				updated_at=now()
			WHERE staging_id=$11`,
			sat, sha, flags["k_sunday"], flags["k_monday"], flags["k_tuesday"], flags["k_wednesday1"],
			flags["k_wednesday2"], flags["k_thursday"], flags["k_friday"], flags["k_saturday"], row.ID)
		if err != nil {
			return processed, review, err
		}
		processed++
	}
	return processed, review, nil
}

func loadDictionary(db *sql.DB) (map[rune]charValues, error) {
	rows, err := db.Query(`
		SELECT COALESCE(s.char_key, sh.char_key), s.sat_value, sh.sha_value
		FROM sat_nums AS s FULL OUTER JOIN sha_nums AS sh ON sh.char_key = s.char_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dictionary := make(map[rune]charValues)
	for rows.Next() {
		var key string
		var sat, sha sql.NullInt64
		if err := rows.Scan(&key, &sat, &sha); err != nil {
			return nil, err
		}
		if utf8.RuneCountInString(key) != 1 {
			return nil, fmt.Errorf("invalid character dictionary key %q", key)
		}
		v := charValues{hasSat: sat.Valid, hasSha: sha.Valid}
		if sat.Valid {
			v.sat = int(sat.Int64)
		}
		if sha.Valid {
			v.sha = int(sha.Int64)
		}
		dictionary[[]rune(key)[0]] = v
	}
	return dictionary, rows.Err()
}

func loadKalakini(db *sql.DB) (map[string]map[rune]bool, error) {
	days := map[string][]string{
		"k_sunday": {"Sunday", "อาทิตย์"}, "k_monday": {"Monday", "จันทร์"}, "k_tuesday": {"Tuesday", "อังคาร"},
		"k_wednesday1": {"Wednesday1", "พุธกลางวัน"}, "k_wednesday2": {"Wednesday2", "พุธกลางคืน"},
		"k_thursday": {"Thursday", "พฤหัสบดี"}, "k_friday": {"Friday", "ศุกร์"}, "k_saturday": {"Saturday", "เสาร์"},
	}
	result := make(map[string]map[rune]bool, len(days))
	for field, names := range days {
		rows, err := db.Query(`SELECT kakis FROM kakis_day WHERE day = $1 OR day_th = $2`, names[0], names[1])
		if err != nil {
			return nil, err
		}
		chars := make(map[rune]bool)
		for rows.Next() {
			var text string
			if err := rows.Scan(&text); err != nil {
				rows.Close()
				return nil, err
			}
			for _, r := range strings.TrimSpace(text) {
				chars[r] = true
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
		result[field] = chars
	}
	return result, nil
}

func calculateSums(name string, dictionary map[rune]charValues) (int, int, []string) {
	sat, sha := 0, 0
	unknown := make(map[rune]bool)
	for _, r := range name {
		v, ok := dictionary[r]
		if !ok {
			unknown[r] = true
			continue
		}
		sat += v.sat
		sha += v.sha
	}
	unknownRunes := make([]string, 0, len(unknown))
	for r := range unknown {
		unknownRunes = append(unknownRunes, string(r))
	}
	sort.Strings(unknownRunes)
	return sat, sha, unknownRunes
}

// formatSumPairs stores the same canonical sliding pairs used by ranking.
// Keeping these legacy columns populated avoids producing structurally
// incomplete rows even though runtime ranking derives pairs from the sums.
func formatSumPairs(total int) []string {
	if total <= 0 {
		return []string{}
	}
	s := strconv.Itoa(total)
	if len(s) == 1 {
		return []string{fmt.Sprintf("%02d", total)}
	}
	if len(s) == 2 {
		return []string{s}
	}
	pairs := make([]string, 0, len(s)-1)
	for i := 0; i < len(s)-1; i++ {
		pairs = append(pairs, s[i:i+2])
	}
	return pairs
}

func calculateKalakini(name string, rules map[string]map[rune]bool) map[string]bool {
	flags := make(map[string]bool, len(rules))
	for field, chars := range rules {
		for _, r := range name {
			if chars[r] {
				flags[field] = true
				break
			}
		}
	}
	return flags
}

func enrichMeaning(ctx context.Context, db *sql.DB, batchID int64, limit int, fallback bool) (processed, review int, err error) {
	condition := "meaning IS NULL AND stage_status = 'staged'"
	if fallback {
		condition = "meaning IS NULL AND stage_status = 'review_required'"
	}
	rows, err := pendingRows(db, batchID, condition, limit)
	if err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}
	results, model, err := requestMeanings(ctx, rows, fallback)
	if err != nil {
		return 0, 0, err
	}
	byName := make(map[string]enrichmentResult, len(results))
	for _, result := range results {
		byName[strings.TrimSpace(result.Name)] = result
	}
	for _, row := range rows {
		result, ok := byName[row.Name]
		if !ok || result.Review || !validMeaning(result.Meaning) || !validGender(result.Gender) {
			if err := markReview(db, row.ID, "meaning_ai_review_required"); err != nil {
				return processed, review, err
			}
			review++
			continue
		}
		_, err := db.Exec(`
			UPDATE data_governance.name_import_staging
			SET meaning=$1, gender=$2, meaning_model=$3, meaning_prompt_version=$4,
				stage_status='meaning_ready', updated_at=now()
			WHERE staging_id=$5`, strings.TrimSpace(result.Meaning), strings.TrimSpace(result.Gender), model, meaningPromptVersion, row.ID)
		if err != nil {
			return processed, review, err
		}
		processed++
	}
	return processed, review, nil
}

func enrichPhonetic(ctx context.Context, db *sql.DB, batchID int64, limit int) (processed, review int, err error) {
	rows, err := pendingRows(db, batchID, "phonetic_version IS NULL", limit)
	if err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}
	results, model, err := requestPhonetics(ctx, rows)
	if err != nil {
		// A malformed model response must not stop a multi-hour import. Retry
		// every name independently; a persistently invalid response is kept for
		// review rather than being silently accepted or retried forever.
		fmt.Printf("phonetic batch response invalid; retrying %d names individually: %v\n", len(rows), err)
		for _, row := range rows {
			single, singleModel, singleErr := requestPhonetics(ctx, []stagedName{row})
			if singleErr != nil || len(single) != 1 {
				if markErr := markReview(db, row.ID, "phonetic_ai_response_invalid"); markErr != nil {
					return processed, review, markErr
				}
				review++
				continue
			}
			result := single[0]
			if result.Review || !validScore(result.Score.Euphony) || !validScore(result.Score.PronunciationEase) || !validScore(result.Score.Rhythm) {
				if markErr := markReview(db, row.ID, "phonetic_ai_review_required"); markErr != nil {
					return processed, review, markErr
				}
				review++
				continue
			}
			_, updateErr := db.Exec(`
				UPDATE data_governance.name_import_staging
				SET euphony_score=$1, pronunciation_ease=$2, rhythm_score=$3,
					phonetic_version=$4, phonetic_model=$5,
					stage_status=CASE WHEN stage_status IN ('meaning_ready', 'review_required') THEN 'phonetic_ready' ELSE stage_status END,
					rejection_reason=NULL,
					updated_at=now()
				WHERE staging_id=$6`, result.Score.Euphony, result.Score.PronunciationEase, result.Score.Rhythm, phoneticVersion, singleModel, row.ID)
			if updateErr != nil {
				return processed, review, updateErr
			}
			processed++
		}
		return processed, review, nil
	}
	byName := make(map[string]phoneticResult, len(results))
	for _, result := range results {
		byName[strings.TrimSpace(result.Name)] = result
	}
	for _, row := range rows {
		result, ok := byName[row.Name]
		if !ok || result.Review || !validScore(result.Score.Euphony) || !validScore(result.Score.PronunciationEase) || !validScore(result.Score.Rhythm) {
			if err := markReview(db, row.ID, "phonetic_ai_review_required"); err != nil {
				return processed, review, err
			}
			review++
			continue
		}
		_, err := db.Exec(`
			UPDATE data_governance.name_import_staging
			SET euphony_score=$1, pronunciation_ease=$2, rhythm_score=$3,
				phonetic_version=$4, phonetic_model=$5,
				stage_status=CASE WHEN stage_status IN ('meaning_ready', 'review_required') THEN 'phonetic_ready' ELSE stage_status END,
				rejection_reason=NULL,
				updated_at=now()
			WHERE staging_id=$6`, result.Score.Euphony, result.Score.PronunciationEase, result.Score.Rhythm, phoneticVersion, model, row.ID)
		if err != nil {
			return processed, review, err
		}
		processed++
	}
	return processed, review, nil
}

func pendingRows(db *sql.DB, batchID int64, condition string, limit int) ([]stagedName, error) {
	query := fmt.Sprintf(`SELECT staging_id, normalized_name FROM data_governance.name_import_staging WHERE batch_id=$1 AND %s ORDER BY staging_id LIMIT $2`, condition)
	rows, err := db.Query(query, batchID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []stagedName
	for rows.Next() {
		var item stagedName
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func markReview(db *sql.DB, id int64, reason string) error {
	_, err := db.Exec(`UPDATE data_governance.name_import_staging SET stage_status='review_required', rejection_reason=$1, updated_at=now() WHERE staging_id=$2`, reason, id)
	return err
}

func requestMeanings(ctx context.Context, names []stagedName, fallback bool) ([]enrichmentResult, string, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil, "", errors.New("OPENAI_API_KEY is not set")
	}
	model := strings.TrimSpace(os.Getenv("NAME_MEANING_MODEL"))
	if model == "" {
		model = defaultMeaningModel
	}
	inputNames := make([]string, 0, len(names))
	for _, item := range names {
		inputNames = append(inputNames, item.Name)
	}
	prompt := `คุณเป็นผู้เชี่ยวชาญภาษาไทย วิเคราะห์รายชื่อไทยตามความหมายจริงอย่างระมัดระวัง
ตอบ JSON object เท่านั้นในรูปแบบ {"results":[{"name":"ชื่อเดิม","meaning":"ความหมายไทยสั้น กระชับ","gender":"ช|ญ|ค","review_required":false}]}
name ต้องตรงกับ input ทุกตัวอักษร, ห้ามแต่งความหมายมงคลเกินข้อเท็จจริง, ห้ามมีภาษาอังกฤษ, ห้ามใส่ชื่อซ้ำใน meaning
หากชื่อไม่ชัดเจน เป็นนามสกุล หรือไม่แน่ใจ ให้ meaning เป็น "" และ review_required=true.
รายชื่อ: ` + strings.Join(inputNames, " | ")
	if fallback {
		prompt = `คุณเป็นบรรณาธิการข้อมูลชื่อไทย ทำงานรอบทบทวนรายชื่อที่ AI ก่อนหน้าไม่มั่นใจ
ตอบ JSON object เท่านั้นในรูปแบบ {"results":[{"name":"ชื่อเดิม","meaning":"ความหมายไทยสั้น กระชับ","gender":"ช|ญ|ค","review_required":false}]}
ใช้รากศัพท์ไทย บาลี หรือสันสกฤตเมื่อมีหลักฐาน ห้ามแต่งความหมายมงคลเกินข้อเท็จจริง และห้ามมีภาษาอังกฤษ
name ต้องตรงกับ input ทุกตัวอักษร. หากแปลรากศัพท์ไม่ได้จริง ให้ meaning เป็น "ชื่อบุคคล" และ gender เป็น "ค"; กรณีนี้ review_required=false.
รายชื่อ: ` + strings.Join(inputNames, " | ")
	}
	body, _ := json.Marshal(map[string]any{
		"model": model, "temperature": 0,
		"response_format": map[string]string{"type": "json_object"},
		"messages":        []map[string]string{{"role": "system", "content": "ตอบ JSON เท่านั้น"}, {"role": "user", "content": prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: aiRequestTimeout()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("OpenAI status %d", resp.StatusCode)
	}
	var chat chatResponse
	if err := json.Unmarshal(raw, &chat); err != nil || len(chat.Choices) == 0 {
		return nil, "", errors.New("invalid OpenAI response")
	}
	var payload struct {
		Results []enrichmentResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(chat.Choices[0].Message.Content), &payload); err != nil {
		return nil, "", fmt.Errorf("parse meaning JSON: %w", err)
	}
	return payload.Results, model, nil
}

func enrichEmbeddings(ctx context.Context, db *sql.DB, batchID int64, limit int) (int, error) {
	rows, err := pendingEmbeddingRows(db, batchID, limit)
	if err != nil || len(rows) == 0 {
		return len(rows), err
	}
	inputs := make([]string, 0, len(rows))
	for _, row := range rows {
		inputs = append(inputs, "ความหมาย: "+row.Meaning)
	}
	vectors, err := requestEmbeddings(ctx, inputs)
	if err != nil {
		return 0, err
	}
	if len(vectors) != len(rows) {
		return 0, fmt.Errorf("embedding response count mismatch: got %d want %d", len(vectors), len(rows))
	}
	for i, row := range rows {
		vectorJSON, err := json.Marshal(vectors[i])
		if err != nil {
			return i, err
		}
		_, err = db.Exec(`
			UPDATE data_governance.name_import_staging
			SET meaning_vector=$1, embedding_model='text-embedding-3-small',
				embedding_input_template='meaning_prefix_th_v1', updated_at=now()
			WHERE staging_id=$2`, string(vectorJSON), row.ID)
		if err != nil {
			return i, err
		}
	}
	return len(rows), nil
}

type embeddingRow struct {
	ID      int64
	Meaning string
}

func pendingEmbeddingRows(db *sql.DB, batchID int64, limit int) ([]embeddingRow, error) {
	rows, err := db.Query(`
		SELECT staging_id, meaning
		FROM data_governance.name_import_staging
		WHERE batch_id=$1 AND meaning IS NOT NULL AND meaning <> '' AND meaning_vector IS NULL
		ORDER BY staging_id LIMIT $2`, batchID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []embeddingRow
	for rows.Next() {
		var row embeddingRow
		if err := rows.Scan(&row.ID, &row.Meaning); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func requestEmbeddings(ctx context.Context, inputs []string) ([][]float64, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}
	body, _ := json.Marshal(map[string]any{"model": "text-embedding-3-small", "input": inputs})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: aiRequestTimeout()}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI embeddings status %d", resp.StatusCode)
	}
	var decoded struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) != len(inputs) {
		return nil, errors.New("invalid embedding response")
	}
	vectors := make([][]float64, len(inputs))
	for _, item := range decoded.Data {
		if item.Index < 0 || item.Index >= len(vectors) || len(item.Embedding) != 1536 {
			return nil, errors.New("unexpected embedding dimensions or index")
		}
		vectors[item.Index] = item.Embedding
	}
	return vectors, nil
}

func promoteCompleteRows(db *sql.DB, batchID int64, limit int) (int, error) {
	type promotionRow struct {
		id              int64
		name            string
		meaning         string
		gender          string
		vector          string
		sat, sha        int
		euphony, ease   int
		rhythm          int
		flags           [8]bool
		phoneticVersion string
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`
		SELECT staging_id, normalized_name, meaning, gender, sat_sum, sha_sum,
			k_sunday, k_monday, k_tuesday, k_wednesday1, k_wednesday2, k_thursday, k_friday, k_saturday,
			euphony_score, pronunciation_ease, rhythm_score, phonetic_version, meaning_vector::text
		FROM data_governance.name_import_staging
		WHERE batch_id=$1 AND promoted_name_id IS NULL
			AND meaning IS NOT NULL AND gender IS NOT NULL AND sat_sum IS NOT NULL AND sha_sum IS NOT NULL
			AND k_sunday IS NOT NULL AND k_monday IS NOT NULL AND k_tuesday IS NOT NULL AND k_wednesday1 IS NOT NULL
			AND k_wednesday2 IS NOT NULL AND k_thursday IS NOT NULL AND k_friday IS NOT NULL AND k_saturday IS NOT NULL
			AND euphony_score IS NOT NULL AND pronunciation_ease IS NOT NULL AND rhythm_score IS NOT NULL AND phonetic_version IS NOT NULL
			AND meaning_vector IS NOT NULL
		ORDER BY staging_id LIMIT $2 FOR UPDATE`, batchID, limit)
	if err != nil {
		return 0, err
	}
	selected := make([]promotionRow, 0, limit)
	for rows.Next() {
		var row promotionRow
		if err := rows.Scan(&row.id, &row.name, &row.meaning, &row.gender, &row.sat, &row.sha, &row.flags[0], &row.flags[1], &row.flags[2], &row.flags[3], &row.flags[4], &row.flags[5], &row.flags[6], &row.flags[7], &row.euphony, &row.ease, &row.rhythm, &row.phoneticVersion, &row.vector); err != nil {
			return 0, err
		}
		selected = append(selected, row)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	// PostgreSQL does not allow another statement on this transaction while
	// the SELECT cursor still has unread protocol messages. Close it before
	// inserting/updating the selected staging rows.
	if err := rows.Close(); err != nil {
		return 0, err
	}

	inserted := 0
	for _, row := range selected {
		var nameID int64
		satPairs := formatSumPairs(row.sat)
		shaPairs := formatSumPairs(row.sha)
		err := tx.QueryRow(`
			INSERT INTO public.names_miracle (
				thname, meaning, gender, sat_sum, sha_sum, satnum, shanum, t_sat, t_sha,
				k_sunday, k_monday, k_tuesday, k_wednesday1, k_wednesday2, k_thursday, k_friday, k_saturday,
				euphony_score, pronunciation_ease, rhythm_score, phonetic_version, meaning_vector)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22::vector)
			ON CONFLICT (thname) DO UPDATE SET thname = EXCLUDED.thname
			RETURNING name_id`, row.name, row.meaning, row.gender, row.sat, row.sha,
			pq.Array(satPairs), pq.Array(shaPairs), pq.Array(satPairs), pq.Array(shaPairs),
			row.flags[0], row.flags[1], row.flags[2], row.flags[3], row.flags[4], row.flags[5], row.flags[6], row.flags[7],
			row.euphony, row.ease, row.rhythm, row.phoneticVersion, row.vector).Scan(&nameID)
		if err != nil {
			return inserted, err
		}
		if _, err := tx.Exec(`UPDATE data_governance.name_import_staging SET promoted_name_id=$1, stage_status='promoted', updated_at=now() WHERE staging_id=$2`, nameID, row.id); err != nil {
			return inserted, err
		}
		inserted++
	}
	if err := tx.Commit(); err != nil {
		return inserted, err
	}
	return inserted, nil
}

func requestPhonetics(ctx context.Context, names []stagedName) ([]phoneticResult, string, error) {
	key := strings.TrimSpace(os.Getenv("TYPHOON_API_KEY"))
	if key == "" {
		return nil, "", errors.New("TYPHOON_API_KEY is not set")
	}
	model := strings.TrimSpace(os.Getenv("TYPHOON_MODEL"))
	if model == "" {
		model = defaultTyphoonModel
	}
	inputNames := make([]string, 0, len(names))
	for _, item := range names {
		inputNames = append(inputNames, item.Name)
	}
	prompt := `ประเมินเฉพาะเสียงชื่อภาษาไทย ไม่วิเคราะห์ความหมายหรือดวง
ตอบ JSON object เท่านั้น: {"results":[{"name":"ชื่อเดิม","score":{"pronunciation_ease":0,"euphony":0,"rhythm":0},"review_required":false}]}
คะแนนเป็นจำนวนเต็ม 0-100, name ต้องตรงกับ input ทุกตัวอักษร
หากไม่มั่นใจ ให้ review_required=true.
รายชื่อ: ` + strings.Join(inputNames, " | ")
	body, _ := json.Marshal(map[string]any{
		"model": model, "temperature": 0.1, "max_tokens": 1800,
		"messages": []map[string]string{{"role": "system", "content": "ตอบ JSON เท่านั้น"}, {"role": "user", "content": prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.opentyphoon.ai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("Typhoon status %d: %s", resp.StatusCode, truncateAPIError(raw, 500))
	}
	var chat chatResponse
	if err := json.Unmarshal(raw, &chat); err != nil || len(chat.Choices) == 0 {
		return nil, "", errors.New("invalid Typhoon response")
	}
	var payload struct {
		Results []phoneticResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(chat.Choices[0].Message.Content)), &payload); err != nil {
		return nil, "", fmt.Errorf("parse phonetic JSON: %w", err)
	}
	return payload.Results, model, nil
}

func truncateAPIError(raw []byte, max int) string {
	message := strings.TrimSpace(string(raw))
	if len(message) > max {
		return message[:max] + "…"
	}
	return message
}

func cleanJSON(content string) string {
	clean := strings.TrimSpace(content)
	if !strings.HasPrefix(clean, "```") {
		return clean
	}
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSpace(strings.TrimPrefix(clean, "json"))
	return strings.TrimSpace(strings.TrimSuffix(clean, "```"))
}

func validMeaning(meaning string) bool {
	return strings.TrimSpace(meaning) != "" && !strings.ContainsAny(meaning, "\n\r")
}
func validGender(gender string) bool { return gender == "ช" || gender == "ญ" || gender == "ค" }
func validScore(score int) bool      { return score >= 0 && score <= 100 }
func fatal(err error)                { fmt.Fprintln(os.Stderr, "name-enrich:", err); os.Exit(1) }

func aiRequestTimeout() time.Duration {
	if raw := strings.TrimSpace(os.Getenv("NAME_AI_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 3 * time.Minute
}
