package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type targetName struct {
	ID     int64
	THName string
}

type aiMeaningResult struct {
	THName  string `json:"thname"`
	Meaning string `json:"meaning"`
	Gender  string `json:"gender"`
}

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("/etc/pj-names-miracle/pj-names-miracle.env")

	host := getEnv("DB_HOST", "43.228.85.200")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "tayap")
	pass := getEnv("DB_PASS", "IntelliP24.X")
	dbname := getEnv("DB_NAME", "tayap")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Connected to database successfully.")

	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set.")
	}

	rows, err := db.Query(`
		SELECT name_id, thname 
		FROM names_miracle 
		WHERE meaning = 'ชื่อบุคคล' 
		   OR meaning IS NULL 
		   OR meaning = ''
		ORDER BY name_id ASC
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	var targets []targetName
	for rows.Next() {
		var t targetName
		if err := rows.Scan(&t.ID, &t.THName); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
		targets = append(targets, t)
	}

	total := len(targets)
	log.Printf("Found %d names with 'ชื่อบุคคล' or empty meaning to enrich.", total)
	if total == 0 {
		log.Println("No names to enrich. Done.")
		return
	}

	httpClient := &http.Client{Timeout: 45 * time.Second}
	batchSize := 12
	updatedCount := 0

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := targets[i:end]

		namesList := make([]string, len(batch))
		for idx, item := range batch {
			namesList[idx] = item.THName
		}

		results, err := requestBatchMeanings(context.Background(), httpClient, apiKey, namesList)
		if err != nil {
			log.Printf("Batch %d-%d meaning generation failed: %v, retrying individually...", i, end, err)
			time.Sleep(2 * time.Second)
		}

		meaningMap := make(map[string]aiMeaningResult)
		for _, res := range results {
			meaningMap[strings.TrimSpace(res.THName)] = res
		}

		for _, item := range batch {
			res, ok := meaningMap[strings.TrimSpace(item.THName)]
			if !ok || res.Meaning == "" {
				// Single fallback if missing
				singleMeaning, singleGender, singleErr := requestSingleMeaning(context.Background(), httpClient, apiKey, item.THName)
				if singleErr != nil {
					log.Printf("Failed single meaning for %s (ID: %d): %v", item.THName, item.ID, singleErr)
					continue
				}
				res = aiMeaningResult{
					THName:  item.THName,
					Meaning: singleMeaning,
					Gender:  singleGender,
				}
			}

			// Generate embedding
			vector, err := requestEmbedding(context.Background(), httpClient, apiKey, "ความหมาย: "+res.Meaning)
			var vectorJSON string
			if err == nil && len(vector) > 0 {
				b, _ := json.Marshal(vector)
				vectorJSON = string(b)
			} else {
				log.Printf("Warning: embedding failed for %s: %v", item.THName, err)
			}

			gender := res.Gender
			if gender != "ช" && gender != "ญ" && gender != "ค" {
				gender = "ค"
			}

			_, err = db.Exec(`
				UPDATE names_miracle 
				SET meaning = $1, 
					gender = $2,
					meaning_vector = CASE WHEN $3 <> '' THEN $3::vector ELSE meaning_vector END
				WHERE name_id = $4
			`, res.Meaning, gender, vectorJSON, item.ID)

			if err != nil {
				log.Printf("Failed to update ID %d (%s): %v", item.ID, item.THName, err)
			} else {
				updatedCount++
			}
		}

		log.Printf("[%d/%d] Enriched %d names successfully (current: %s)", updatedCount, total, len(batch), namesList[0])
		time.Sleep(300 * time.Millisecond)
	}

	log.Printf("Completed! Successfully enriched %d of %d names.", updatedCount, total)
}

func requestBatchMeanings(ctx context.Context, client *http.Client, apiKey string, names []string) ([]aiMeaningResult, error) {
	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญภาษาไทยและศัพทวิทยา
โปรดวิเคราะห์และสร้างคำแปลความหมายมงคลที่สละสลวย ไพเราะ สื่อความหมายเชิงบวก ไม่เกิน 15 คำ ไม่ใส่คำว่า "หมายถึง"
สำหรับรายชื่อภาษาไทยต่อไปนี้: %s
ตอบเป็น JSON object เท่านั้นในรูปแบบ:
{"results": [
  {"thname": "...", "meaning": "คำแปลสั้นกระชับ สละสลวย", "gender": "ช|ญ|ค"}
]}`, strings.Join(names, ", "))

	body, _ := json.Marshal(map[string]any{
		"model":           "gpt-4o-mini",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": "คุณเป็นผู้เชี่ยวชาญภาษาไทยที่ตอบเฉพาะ JSON เท่านั้น"},
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI API status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("parse choices: %w", err)
	}

	var res struct {
		Results []aiMeaningResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &res); err != nil {
		return nil, fmt.Errorf("parse results JSON: %w", err)
	}

	return res.Results, nil
}

func requestSingleMeaning(ctx context.Context, client *http.Client, apiKey string, name string) (string, string, error) {
	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญภาษาไทยและศัพทวิทยา
โปรดวิเคราะห์ความหมายมงคลของชื่อ "%s" ให้สั้น กระชับ สละสลวย ไม่เกิน 15 คำ ไม่ใส่คำว่า "หมายถึง"
ตอบเป็น JSON object เท่านั้น:
{"meaning": "...", "gender": "ช|ญ|ค"}`, name)

	body, _ := json.Marshal(map[string]any{
		"model":           "gpt-4o-mini",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": "คุณเป็นผู้เชี่ยวชาญภาษาไทยที่ตอบเฉพาะ JSON เท่านั้น"},
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("OpenAI API status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return "", "", fmt.Errorf("parse choices: %w", err)
	}

	var res struct {
		Meaning string `json:"meaning"`
		Gender  string `json:"gender"`
	}
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &res); err != nil {
		return "", "", err
	}
	return strings.TrimSpace(res.Meaning), strings.TrimSpace(res.Gender), nil
}

func requestEmbedding(ctx context.Context, client *http.Client, apiKey string, text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]any{
		"model": "text-embedding-3-small",
		"input": text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI embedding status %d: %s", resp.StatusCode, string(raw))
	}

	var res struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil || len(res.Data) == 0 {
		return nil, fmt.Errorf("parse embedding: %w", err)
	}
	return res.Data[0].Embedding, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
