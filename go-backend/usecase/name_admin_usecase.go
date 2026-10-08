package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go-backend/domain"
	"go-backend/repository"
)

type NameAdminUseCase interface {
	AnalyzeName(ctx context.Context, thname string, autoMeaning bool, gender string) (*domain.NameAnalysisResult, error)
	SaveName(ctx context.Context, req *domain.CreateAdminNameRequest) (*domain.AdminNameItem, error)
	ListNames(ctx context.Context, search string, page int, pageSize int) ([]domain.AdminNameItem, int, error)
	GetTotalNamesCount(ctx context.Context) (int, error)
}

type nameAdminUseCase struct {
	repo       repository.NameAdminRepository
	httpClient *http.Client
}

func NewNameAdminUseCase(repo repository.NameAdminRepository) NameAdminUseCase {
	return &nameAdminUseCase{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (u *nameAdminUseCase) AnalyzeName(ctx context.Context, thname string, autoMeaning bool, gender string) (*domain.NameAnalysisResult, error) {
	trimmed := strings.TrimSpace(thname)
	if trimmed == "" {
		return nil, errors.New("กรุณาระบุชื่อภาษาไทย")
	}

	exists, existingID, err := u.repo.ExistsTHName(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("check existing name: %w", err)
	}

	satDict, shaDict, err := u.repo.GetDictionary(ctx)
	if err != nil {
		return nil, fmt.Errorf("load dictionary: %w", err)
	}

	kakisRules, err := u.repo.GetKalakiniRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load kalakini rules: %w", err)
	}

	satSum := 0
	shaSum := 0
	for _, r := range trimmed {
		if val, ok := satDict[r]; ok {
			satSum += val
		}
		if val, ok := shaDict[r]; ok {
			shaSum += val
		}
	}

	satPairs := formatSumPairs(satSum)
	shaPairs := formatSumPairs(shaSum)

	dayDefs := []struct {
		field string
		label string
	}{
		{"k_sunday", "วันอาทิตย์"},
		{"k_monday", "วันจันทร์"},
		{"k_tuesday", "วันอังคาร"},
		{"k_wednesday1", "วันพุธ (กลางวัน)"},
		{"k_wednesday2", "วันพุธ (กลางคืน)"},
		{"k_thursday", "วันพฤหัสบดี"},
		{"k_friday", "วันศุกร์"},
		{"k_saturday", "วันเสาร์"},
	}

	daysStatus := make([]domain.DayKalakiniStatus, 0, len(dayDefs))
	for _, d := range dayDefs {
		isKakis := false
		if chars, ok := kakisRules[d.field]; ok {
			for _, r := range trimmed {
				if chars[r] {
					isKakis = true
					break
				}
			}
		}
		daysStatus = append(daysStatus, domain.DayKalakiniStatus{
			Field:   d.field,
			DayTH:   d.label,
			IsKakis: isKakis,
		})
	}

	result := &domain.NameAnalysisResult{
		THName:            trimmed,
		SatSum:            satSum,
		ShaSum:            shaSum,
		SatPairs:          satPairs,
		ShaPairs:          shaPairs,
		Days:              daysStatus,
		Gender:            gender,
		AlreadyExists:     exists,
		ExistingID:        existingID,
		EuphonyScore:      80,
		PronunciationEase: 80,
		RhythmScore:       80,
	}
	if result.Gender == "" {
		result.Gender = "ค"
	}

	if autoMeaning {
		meaning, suggestedGender, err := u.generateAIMeaning(ctx, trimmed)
		if err == nil && meaning != "" {
			result.Meaning = meaning
			if suggestedGender != "" && (result.Gender == "" || result.Gender == "ค") {
				result.Gender = suggestedGender
			}
		}

		euphony, ease, rhythm, err := u.scorePhonetics(ctx, trimmed)
		if err == nil && euphony > 0 {
			result.EuphonyScore = euphony
			result.PronunciationEase = ease
			result.RhythmScore = rhythm
		}
	}

	return result, nil
}

func (u *nameAdminUseCase) SaveName(ctx context.Context, req *domain.CreateAdminNameRequest) (*domain.AdminNameItem, error) {
	trimmed := strings.TrimSpace(req.THName)
	if trimmed == "" {
		return nil, errors.New("กรุณาระบุชื่อภาษาไทย")
	}

	exists, _, err := u.repo.ExistsTHName(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("check existing name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("ชื่อ %q มีอยู่ในระบบอยู่แล้ว", trimmed)
	}

	satDict, shaDict, err := u.repo.GetDictionary(ctx)
	if err != nil {
		return nil, fmt.Errorf("load dictionary: %w", err)
	}
	kakisRules, err := u.repo.GetKalakiniRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load kalakini rules: %w", err)
	}

	satSum := 0
	shaSum := 0
	for _, r := range trimmed {
		if val, ok := satDict[r]; ok {
			satSum += val
		}
		if val, ok := shaDict[r]; ok {
			shaSum += val
		}
	}

	satPairs := formatSumPairs(satSum)
	shaPairs := formatSumPairs(shaSum)

	kakisFlags := make(map[string]bool)
	for field, chars := range kakisRules {
		for _, r := range trimmed {
			if chars[r] {
				kakisFlags[field] = true
				break
			}
		}
	}

	gender := strings.TrimSpace(req.Gender)
	if gender != "ช" && gender != "หญิง" && gender != "ชาย" && gender != "ญ" && gender != "ทุกเพศ" {
		gender = "ค"
	}
	if gender == "ชาย" {
		gender = "ช"
	}
	if gender == "หญิง" {
		gender = "ญ"
	}
	if gender == "ทุกเพศ" {
		gender = "ค"
	}

	meaning := strings.TrimSpace(req.Meaning)
	var vectorJSON string
	if meaning != "" {
		vector, err := u.generateEmbedding(ctx, "ความหมาย: "+meaning)
		if err == nil && len(vector) > 0 {
			b, _ := json.Marshal(vector)
			vectorJSON = string(b)
		}
	}

	euphony := req.EuphonyScore
	if euphony <= 0 {
		euphony = 80
	}
	ease := req.PronunciationEase
	if ease <= 0 {
		ease = 80
	}
	rhythm := req.RhythmScore
	if rhythm <= 0 {
		rhythm = 80
	}

	item := &domain.AdminNameItem{
		THName:            trimmed,
		Meaning:           meaning,
		Gender:            gender,
		SatSum:            satSum,
		ShaSum:            shaSum,
		SatPairs:          satPairs,
		ShaPairs:          shaPairs,
		KSunday:           kakisFlags["k_sunday"],
		KMonday:           kakisFlags["k_monday"],
		KTuesday:          kakisFlags["k_tuesday"],
		KWednesday1:       kakisFlags["k_wednesday1"],
		KWednesday2:       kakisFlags["k_wednesday2"],
		KThursday:         kakisFlags["k_thursday"],
		KFriday:           kakisFlags["k_friday"],
		KSaturday:         kakisFlags["k_saturday"],
		EuphonyScore:      euphony,
		PronunciationEase: ease,
		RhythmScore:       rhythm,
		PhoneticVersion:   "admin_v1",
		HasVector:         vectorJSON != "",
	}

	nameID, err := u.repo.InsertCompleteName(ctx, item, vectorJSON)
	if err != nil {
		return nil, fmt.Errorf("insert name: %w", err)
	}
	item.NameID = nameID
	return item, nil
}

func (u *nameAdminUseCase) ListNames(ctx context.Context, search string, page int, pageSize int) ([]domain.AdminNameItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return u.repo.GetRecentNames(ctx, search, pageSize, offset)
}

func (u *nameAdminUseCase) GetTotalNamesCount(ctx context.Context) (int, error) {
	return u.repo.GetTotalNamesCount(ctx)
}

// formatSumPairs formats sliding 2-digit pairs from a sum (e.g. 16 -> ["16"], 111 -> ["11", "11"]).
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

func (u *nameAdminUseCase) generateAIMeaning(ctx context.Context, name string) (string, string, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("TYPHOON_API_KEY"))
		if key != "" {
			return u.generateAIMeaningTyphoon(ctx, name, key)
		}
		return "", "", errors.New("neither OPENAI_API_KEY nor TYPHOON_API_KEY is configured")
	}

	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญภาษาไทยและศัพทวิทยา
โปรดวิเคราะห์ความหมายของชื่อ "%s" ตามรากศัพท์จริง (บาลี สันสกฤต หรือภาษาไทย)
ตอบเป็น JSON object เท่านั้น:
{"meaning": "ความหมายภาษาไทยที่สั้น กระชับ สละสลวย ไม่เกิน 15 คำ ไม่ต้องใส่คำว่า หมายถึง", "gender": "ช|ญ|ค"}
ข้อกำหนด:
- ห้ามแต่งความหมายเกินข้อเท็จจริง
- หากเป็นคำไทยแท้ให้แปลตามความหมายตรง
- gender ให้ระบุ "ช" (ชาย), "ญ" (หญิง), หรือ "ค" (ทุกเพศ)`, name)

	body, _ := json.Marshal(map[string]any{
		"model": "gpt-4o-mini",
		"temperature": 0.2,
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
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(raw))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return "", "", errors.New("failed to parse OpenAI response")
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

func (u *nameAdminUseCase) generateAIMeaningTyphoon(ctx context.Context, name string, key string) (string, string, error) {
	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญภาษาไทย โปรดวิเคราะห์ความหมายของชื่อ "%s" ตามรากศัพท์จริง สั้น กระชับ สละสลวย
ตอบเป็น JSON: {"meaning": "คำแปลสั้นไม่เกิน 15 คำ", "gender": "ช|ญ|ค"}`, name)

	body, _ := json.Marshal(map[string]any{
		"model": "typhoon-v2.5-30b-a3b-instruct",
		"temperature": 0.2,
		"messages": []map[string]string{
			{"role": "system", "content": "ตอบเฉพาะ JSON object เท่านั้น"},
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.opentyphoon.ai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("Typhoon API error %d: %s", resp.StatusCode, string(raw))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return "", "", errors.New("failed to parse Typhoon response")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var res struct {
		Meaning string `json:"meaning"`
		Gender  string `json:"gender"`
	}
	if err := json.Unmarshal([]byte(content), &res); err != nil {
		return "", "", err
	}
	return strings.TrimSpace(res.Meaning), strings.TrimSpace(res.Gender), nil
}

func (u *nameAdminUseCase) scorePhonetics(ctx context.Context, name string) (int, int, int, error) {
	key := strings.TrimSpace(os.Getenv("TYPHOON_API_KEY"))
	if key == "" {
		return 80, 80, 80, nil
	}

	prompt := fmt.Sprintf(`ประเมินเฉพาะเสียงชื่อภาษาไทย "%s" ไม่วิเคราะห์ความหมาย
ตอบเป็น JSON: {"euphony": 0-100, "pronunciation_ease": 0-100, "rhythm": 0-100}`, name)

	body, _ := json.Marshal(map[string]any{
		"model": "typhoon-v2.5-30b-a3b-instruct",
		"temperature": 0.1,
		"messages": []map[string]string{
			{"role": "system", "content": "ตอบเฉพาะ JSON object เท่านั้น"},
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.opentyphoon.ai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return 80, 80, 80, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return 80, 80, 80, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 80, 80, 80, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 80, 80, 80, nil
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		return 80, 80, 80, nil
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var res struct {
		Euphony           int `json:"euphony"`
		PronunciationEase int `json:"pronunciation_ease"`
		Rhythm            int `json:"rhythm"`
	}
	if err := json.Unmarshal([]byte(content), &res); err != nil {
		return 80, 80, 80, nil
	}

	if res.Euphony <= 0 {
		res.Euphony = 80
	}
	if res.PronunciationEase <= 0 {
		res.PronunciationEase = 80
	}
	if res.Rhythm <= 0 {
		res.Rhythm = 80
	}
	return res.Euphony, res.PronunciationEase, res.Rhythm, nil
}

func (u *nameAdminUseCase) generateEmbedding(ctx context.Context, text string) ([]float64, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	body, _ := json.Marshal(map[string]any{
		"model": "text-embedding-3-small",
		"input": text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI embeddings error %d: %s", resp.StatusCode, string(raw))
	}

	var res struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &res); err != nil || len(res.Data) == 0 {
		return nil, errors.New("failed to parse embedding vector")
	}
	return res.Data[0].Embedding, nil
}
