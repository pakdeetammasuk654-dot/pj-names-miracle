package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KranaxALT/thailunar"
	"go-backend/domain"
)

const (
	maxAuspiciousSearchDays = 366
	maxRecommendations      = 3
)

type AuspiciousService struct {
	repo        domain.AuspiciousRuleRepository
	contentRepo domain.AuspiciousContentRepository
	clock       func() time.Time
}

func NewAuspiciousService(repo domain.AuspiciousRuleRepository, contentRepos ...domain.AuspiciousContentRepository) *AuspiciousService {
	service := &AuspiciousService{repo: repo, clock: time.Now}
	if len(contentRepos) > 0 {
		service.contentRepo = contentRepos[0]
	}
	return service
}

func (s *AuspiciousService) ListActivities(ctx context.Context) ([]domain.AuspiciousActivity, error) {
	return s.repo.ListActivities(ctx)
}

func (s *AuspiciousService) CalculateTaksa(ctx context.Context, birthDate, birthTime string, currentTime time.Time) (TaksaCalculationResult, error) {
	texts := map[string]string{}
	if s.contentRepo != nil {
		var err error
		texts, err = s.contentRepo.GetTextMap(ctx)
		if err != nil {
			return TaksaCalculationResult{}, err
		}
	}
	return CalculateTaksaWithTexts(birthDate, birthTime, currentTime, texts), nil
}

type auspiciousFacts struct {
	weekday            string
	lunarMonth         int
	lunarPhase         string
	lunarDay           int
	isWanPhra          bool
	isWeekend          bool
	monkCeremony       bool
	isKalaYokThongChai bool
	isKalaYokAthibodi  bool
	isKalaYokUbat      bool
	isKalaYokLokawinat bool
}

func (s *AuspiciousService) Search(ctx context.Context, req domain.AuspiciousSearchRequest) (domain.AuspiciousSearchResponse, error) {
	activities, err := s.repo.ListActivities(ctx)
	if err != nil {
		return domain.AuspiciousSearchResponse{}, err
	}
	var activity domain.AuspiciousActivity
	for _, candidate := range activities {
		if candidate.Code == req.ActivityCode {
			activity = candidate
			break
		}
	}
	if activity.Code == "" {
		return domain.AuspiciousSearchResponse{}, errors.New("ไม่พบประเภทงานที่ต้องการหาฤกษ์")
	}

	bangkok, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return domain.AuspiciousSearchResponse{}, fmt.Errorf("load Bangkok timezone: %w", err)
	}
	start, err := time.ParseInLocation("2006-01-02", req.StartDate, bangkok)
	if err != nil {
		return domain.AuspiciousSearchResponse{}, errors.New("วันที่เริ่มต้นไม่ถูกต้อง")
	}
	end, err := time.ParseInLocation("2006-01-02", req.EndDate, bangkok)
	if err != nil {
		return domain.AuspiciousSearchResponse{}, errors.New("วันที่สิ้นสุดไม่ถูกต้อง")
	}
	if end.Before(start) {
		return domain.AuspiciousSearchResponse{}, errors.New("วันที่สิ้นสุดต้องไม่มาก่อนวันที่เริ่มต้น")
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days > maxAuspiciousSearchDays {
		return domain.AuspiciousSearchResponse{}, fmt.Errorf("เลือกช่วงเวลาได้ไม่เกิน %d วัน", maxAuspiciousSearchDays)
	}

	if req.BirthTime != "" && req.BirthDate == "" {
		return domain.AuspiciousSearchResponse{}, errors.New("กรุณาระบุวันเกิดก่อนระบุเวลาเกิด")
	}
	if req.BirthDate != "" && req.BirthTime == "" {
		return domain.AuspiciousSearchResponse{}, errors.New("กรุณาระบุเวลาเกิดเพื่อคำนวณทักษาจร (ใช้เวลา 06:00 น. เป็นเวลาข้ามวัน)")
	}
	if req.BirthTime != "" {
		birthTime, timeErr := time.Parse("15:04", req.BirthTime)
		if timeErr != nil || birthTime.Format("15:04") != req.BirthTime {
			return domain.AuspiciousSearchResponse{}, errors.New("เวลาเกิดต้องอยู่ในรูปแบบ HH:MM")
		}
	}
	if req.DayOfBirth != "" && getStartPlanet(req.DayOfBirth) == 0 {
		return domain.AuspiciousSearchResponse{}, errors.New("วันเกิดทางโหราศาสตร์ไม่ถูกต้อง")
	}
	if req.Age < 0 {
		return domain.AuspiciousSearchResponse{}, errors.New("อายุย่างต้องไม่ติดลบ")
	}
	if req.BirthDate != "" {
		bDate, bErr := time.ParseInLocation("2006-01-02", req.BirthDate, bangkok)
		if bErr != nil {
			return domain.AuspiciousSearchResponse{}, errors.New("วันเกิดไม่ถูกต้อง")
		}
		currentDate := s.clock().In(bangkok)
		if bDate.After(currentDate) {
			return domain.AuspiciousSearchResponse{}, errors.New("วันเกิดต้องไม่อยู่ในอนาคต")
		}
		if req.DayOfBirth == "" {
			dayCode, _ := DetermineAstrologicalBirthDay(bDate, req.BirthTime)
			req.DayOfBirth = dayCode
		}
		_, req.Age = calculateCurrentAges(bDate, currentDate)
	}

	rules, err := s.repo.ListRules(ctx, req.ActivityCode)
	if err != nil {
		return domain.AuspiciousSearchResponse{}, err
	}
	if len(rules) == 0 {
		return domain.AuspiciousSearchResponse{}, errors.New("ยังไม่มีกฎฤกษ์สำหรับกิจกรรมนี้")
	}
	texts := map[string]string{}
	if s.contentRepo != nil {
		texts, err = s.contentRepo.GetTextMap(ctx)
		if err != nil {
			return domain.AuspiciousSearchResponse{}, err
		}
	}

	results := make([]domain.AuspiciousDateResult, 0, days)
	excluded := 0
	ruleSetSeen := make(map[string]bool)
	ruleSets := make([]string, 0)
	for _, rule := range rules {
		if !ruleSetSeen[rule.RuleSetName] {
			ruleSetSeen[rule.RuleSetName] = true
			ruleSets = append(ruleSets, rule.RuleSetName)
		}
	}

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		lunar, conversionErr := thailunar.FromGregorian(date.Year(), int(date.Month()), date.Day())
		if conversionErr != nil {
			return domain.AuspiciousSearchResponse{}, conversionErr
		}
		facts := factsForDate(date, lunar, req.MonkCeremony)
		result, blocked, evaluationErr := evaluateAuspiciousDateWithTexts(date, lunar, facts, rules, req, texts)
		if evaluationErr != nil {
			return domain.AuspiciousSearchResponse{}, evaluationErr
		}
		if blocked {
			excluded++
			continue
		}
		result.OutfitAdvice = buildOutfitColorAdvice(date, req.DayOfBirth, req.Age)
		results = append(results, result)
	}

	sort.SliceStable(results, func(i, j int) bool {
		if req.ActivityCode == "wedding" {
			iHasRiangMon := hasAuspiciousRuleCode(results[i], "thai_wedding_riang_mon")
			jHasRiangMon := hasAuspiciousRuleCode(results[j], "thai_wedding_riang_mon")
			if iHasRiangMon != jHasRiangMon {
				return iHasRiangMon
			}
		}
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Date < results[j].Date
	})
	if len(results) > maxRecommendations {
		results = results[:maxRecommendations]
	}

	return domain.AuspiciousSearchResponse{
		Activity:        activity,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		EvaluatedDays:   days,
		ExcludedDays:    excluded,
		Recommendations: results,
		RuleSets:        ruleSets,
		Disclaimer:      "",
		GeneratedAt:     s.clock().In(bangkok),
	}, nil
}

func factsForDate(date time.Time, lunar thailunar.Date, monkCeremony bool) auspiciousFacts {
	weekday := [...]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}[date.Weekday()]
	phase := "waxing"
	if lunar.Phase() == thailunar.Waning {
		phase = "waning"
	}
	kalaYok := kalaYokForDate(date)
	return auspiciousFacts{
		weekday:            weekday,
		lunarMonth:         lunar.Month(),
		lunarPhase:         phase,
		lunarDay:           lunar.PhaseDay(),
		isWanPhra:          lunar.IsWanPhra(),
		isWeekend:          date.Weekday() == time.Saturday || date.Weekday() == time.Sunday,
		monkCeremony:       monkCeremony,
		isKalaYokThongChai: kalaYok.thongChai,
		isKalaYokAthibodi:  kalaYok.athibodi,
		isKalaYokUbat:      kalaYok.ubat,
		isKalaYokLokawinat: kalaYok.lokawinat,
	}
}

func evaluateAuspiciousDate(date time.Time, lunar thailunar.Date, facts auspiciousFacts, rules []domain.AuspiciousRule, req domain.AuspiciousSearchRequest) (domain.AuspiciousDateResult, bool, error) {
	return evaluateAuspiciousDateWithTexts(date, lunar, facts, rules, req, nil)
}

func evaluateAuspiciousDateWithTexts(date time.Time, lunar thailunar.Date, facts auspiciousFacts, rules []domain.AuspiciousRule, req domain.AuspiciousSearchRequest, texts map[string]string) (domain.AuspiciousDateResult, bool, error) {
	matchedScores := make(map[string]domain.AuspiciousRule)
	blocked := false
	for _, rule := range rules {
		matches, err := evaluateExpression(rule.Condition, facts)
		if err != nil {
			return domain.AuspiciousDateResult{}, false, fmt.Errorf("evaluate rule %s: %w", rule.Code, err)
		}
		if !matches {
			continue
		}
		if rule.Effect == "exclude" {
			blocked = true
			continue
		}
		current, exists := matchedScores[rule.SignalGroup]
		if !exists || rule.Priority > current.Priority || (rule.Priority == current.Priority && math.Abs(float64(rule.Score)) > math.Abs(float64(current.Score))) {
			matchedScores[rule.SignalGroup] = rule
		}
	}
	if req.AvoidWanPhra && facts.isWanPhra {
		blocked = true
	}
	if blocked {
		return domain.AuspiciousDateResult{}, true, nil
	}

	score := 50
	reasons := make([]domain.AuspiciousReason, 0, len(matchedScores)+1)
	selected := make([]domain.AuspiciousRule, 0, len(matchedScores))
	for _, rule := range matchedScores {
		selected = append(selected, rule)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Score > selected[j].Score })
	for _, rule := range selected {
		score += rule.Score
		impact := "support"
		if rule.Score < 0 {
			impact = "caution"
		}
		reasons = append(reasons, domain.AuspiciousReason{RuleCode: rule.Code, Name: rule.NameTH, Explanation: rule.DescriptionTH, Meaning: rule.MeaningTH, Impact: impact, Score: rule.Score})
	}
	if req.PreferWeekend && facts.isWeekend {
		score += 8
		reasons = append(reasons, domain.AuspiciousReason{RuleCode: "user_prefer_weekend", Name: "ตรงวันหยุด", Explanation: "ตรงกับความสะดวกที่เลือกไว้ว่าเน้นวันเสาร์–อาทิตย์", Impact: "practical", Score: 8})
	}
	if req.DayOfBirth != "" {
		if taksa := EvaluateDayTaksaWithTexts(req.DayOfBirth, req.Age, date.Weekday(), texts); taksa != nil {
			score += taksa.Score
			reasons = append(reasons, domain.AuspiciousReason{
				RuleCode:    "thaksa_" + strings.ToLower(req.DayOfBirth) + "_" + date.Weekday().String(),
				Name:        taksa.Name,
				Explanation: taksa.Explanation,
				Meaning:     taksa.Meaning,
				Impact:      taksa.Impact,
				Score:       taksa.Score,
			})
		}
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	if len(reasons) == 0 {
		reasons = append(reasons, domain.AuspiciousReason{RuleCode: "no_blocking_rule", Name: "ไม่พบข้อห้าม", Explanation: "วันนี้ไม่ตรงกับเงื่อนไขต้องห้ามในชุดกฎที่เลือก", Impact: "neutral", Score: 0})
	}

	return domain.AuspiciousDateResult{
		Date:         date.Format("2006-01-02"),
		ThaiDate:     thaiLongDate(date),
		Weekday:      lunar.WeekdayName(),
		LunarLabel:   lunar.Phase().String() + " " + strconv.Itoa(lunar.PhaseDay()) + " ค่ำ เดือน " + lunar.MonthName(),
		LunarPhase:   lunar.Phase().String(),
		LunarDay:     lunar.PhaseDay(),
		LunarMonth:   lunar.Month(),
		IsWanPhra:    lunar.IsWanPhra(),
		ImportantDay: importantBuddhistDay(lunar),
		IsWeekend:    facts.isWeekend,
		Score:        score,
		Rating:       auspiciousRating(score),
		Reasons:      reasons,
	}, false, nil
}

func importantBuddhistDay(lunar thailunar.Date) string {
	month, day := lunar.Month(), lunar.PhaseDay()
	phase := lunar.Phase()
	if phase == thailunar.Waxing && day == 15 {
		switch month {
		case 3:
			return "วันมาฆบูชา"
		case 6:
			return "วันวิสาขบูชา"
		case 8:
			return "วันอาสาฬหบูชา"
		case 11:
			return "วันออกพรรษา"
		}
	}
	if phase == thailunar.Waning && day == 1 && month == 8 {
		return "วันเข้าพรรษา"
	}
	return ""
}

func evaluateExpression(expression domain.RuleExpression, facts auspiciousFacts) (bool, error) {
	if len(expression.All) > 0 {
		for _, child := range expression.All {
			matched, err := evaluateExpression(child, facts)
			if err != nil || !matched {
				return false, err
			}
		}
		return true, nil
	}
	if len(expression.Any) > 0 {
		for _, child := range expression.Any {
			matched, err := evaluateExpression(child, facts)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	if expression.Not != nil {
		matched, err := evaluateExpression(*expression.Not, facts)
		return !matched, err
	}

	actual, ok := factValue(expression.Field, facts)
	if !ok {
		return false, fmt.Errorf("unsupported field %q", expression.Field)
	}
	switch expression.Operator {
	case "eq":
		return rawEquals(actual, expression.Value)
	case "in":
		var values []any
		if err := json.Unmarshal(expression.Value, &values); err != nil {
			return false, err
		}
		for _, expected := range values {
			if normalizedEqual(actual, expected) {
				return true, nil
			}
		}
		return false, nil
	case "equals_mapped":
		mapKey, ok := factValue(expression.MapField, facts)
		if !ok {
			return false, fmt.Errorf("unsupported map field %q", expression.MapField)
		}
		var mapping map[string]any
		if err := json.Unmarshal(expression.Value, &mapping); err != nil {
			return false, err
		}
		expected, exists := mapping[fmt.Sprint(mapKey)]
		return exists && normalizedEqual(actual, expected), nil
	default:
		return false, fmt.Errorf("unsupported operator %q", expression.Operator)
	}
}

func factValue(field string, facts auspiciousFacts) (any, bool) {
	switch field {
	case "weekday":
		return facts.weekday, true
	case "lunar_month":
		return facts.lunarMonth, true
	case "lunar_phase":
		return facts.lunarPhase, true
	case "lunar_day":
		return facts.lunarDay, true
	case "is_wan_phra":
		return facts.isWanPhra, true
	case "is_weekend":
		return facts.isWeekend, true
	case "monk_ceremony":
		return facts.monkCeremony, true
	case "is_kala_yok_thongchai":
		return facts.isKalaYokThongChai, true
	case "is_kala_yok_athibodi":
		return facts.isKalaYokAthibodi, true
	case "is_kala_yok_ubat":
		return facts.isKalaYokUbat, true
	case "is_kala_yok_lokawinat":
		return facts.isKalaYokLokawinat, true
	default:
		return nil, false
	}
}

func rawEquals(actual any, raw json.RawMessage) (bool, error) {
	var expected any
	if err := json.Unmarshal(raw, &expected); err != nil {
		return false, err
	}
	return normalizedEqual(actual, expected), nil
}

func normalizedEqual(actual, expected any) bool {
	switch value := actual.(type) {
	case int:
		number, ok := expected.(float64)
		return ok && value == int(number)
	case bool:
		other, ok := expected.(bool)
		return ok && value == other
	case string:
		other, ok := expected.(string)
		return ok && value == other
	default:
		return fmt.Sprint(actual) == fmt.Sprint(expected)
	}
}

func auspiciousRating(score int) string {
	switch {
	case score >= 85:
		return "โดดเด่นที่สุด"
	case score >= 70:
		return "ดีมาก"
	case score >= 58:
		return "ดี"
	default:
		return "ใช้เป็นทางเลือก"
	}
}

var thaiMonthNames = [...]string{"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}

func thaiLongDate(date time.Time) string {
	return strconv.Itoa(date.Day()) + " " + thaiMonthNames[int(date.Month())-1] + " " + strconv.Itoa(date.Year()+543)
}

func hasAuspiciousRuleCode(result domain.AuspiciousDateResult, code string) bool {
	for _, reason := range result.Reasons {
		if reason.RuleCode == code {
			return true
		}
	}
	return false
}
