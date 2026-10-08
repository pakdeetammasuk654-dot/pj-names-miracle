package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/KranaxALT/thailunar"
	"go-backend/domain"
)

func TestAuspiciousHardExclusionWinsPositiveRule(t *testing.T) {
	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	lunar, err := thailunar.FromGregorian(date.Year(), int(date.Month()), date.Day())
	if err != nil {
		t.Fatal(err)
	}
	facts := factsForDate(date, lunar, false)
	rules := []domain.AuspiciousRule{
		{Code: "good", SignalGroup: "good", Effect: "score", Score: 30, Priority: 10, Condition: eqExpression("lunar_day", lunar.PhaseDay())},
		{Code: "blocked", SignalGroup: "blocked", Effect: "exclude", Priority: 1000, Condition: eqExpression("lunar_day", lunar.PhaseDay())},
	}

	_, blocked, err := evaluateAuspiciousDate(date, lunar, facts, rules, domain.AuspiciousSearchRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !blocked {
		t.Fatal("hard exclusion must win even when a positive rule also matches")
	}
}

func TestAuspiciousScoreCountsOneRulePerSignalGroup(t *testing.T) {
	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	lunar, err := thailunar.FromGregorian(date.Year(), int(date.Month()), date.Day())
	if err != nil {
		t.Fatal(err)
	}
	facts := factsForDate(date, lunar, false)
	condition := eqExpression("lunar_day", lunar.PhaseDay())
	rules := []domain.AuspiciousRule{
		{Code: "alias-a", SignalGroup: "maha_chok", NameTH: "กฎ A", Effect: "score", Score: 12, Priority: 100, Condition: condition},
		{Code: "alias-b", SignalGroup: "maha_chok", NameTH: "กฎ B", DescriptionTH: "เหตุผลที่ตรงกฎ", MeaningTH: "ความหมายจากฐานข้อมูล", Effect: "score", Score: 18, Priority: 200, Condition: condition},
	}

	result, blocked, err := evaluateAuspiciousDate(date, lunar, facts, rules, domain.AuspiciousSearchRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if blocked {
		t.Fatal("date should not be blocked")
	}
	if result.Score != 68 {
		t.Fatalf("score = %d, want 68 (base 50 + highest-priority group score 18)", result.Score)
	}
	if len(result.Reasons) != 1 || result.Reasons[0].RuleCode != "alias-b" {
		t.Fatalf("reasons = %#v, want only alias-b", result.Reasons)
	}
	if result.Reasons[0].Meaning != "ความหมายจากฐานข้อมูล" {
		t.Fatalf("meaning = %q, want database-backed rule meaning", result.Reasons[0].Meaning)
	}
}

func TestKalaYokForDateUsesThaiAstrologicalYearBoundary(t *testing.T) {
	beforeNewYear := kalaYokForDate(time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC))
	if !beforeNewYear.thongChai || !beforeNewYear.athibodi {
		t.Fatalf("10 April 2026 = %#v, want old-year Thong Chai and Athibodi", beforeNewYear)
	}

	afterNewYear := kalaYokForDate(time.Date(2026, time.April, 17, 0, 0, 0, 0, time.UTC))
	if afterNewYear.thongChai || afterNewYear.athibodi || afterNewYear.ubat || afterNewYear.lokawinat {
		t.Fatalf("17 April 2026 = %#v, want no new-year Kala Yok day signal", afterNewYear)
	}
}

func TestKalaYokForDateSupportsOverlappingSignals(t *testing.T) {
	// In the 2569/2026 Kala Yok year, Monday is both Thong Chai and Lokawinat.
	kalaYok := kalaYokForDate(time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC))
	if !kalaYok.thongChai || !kalaYok.lokawinat {
		t.Fatalf("31 August 2026 = %#v, want Thong Chai and Lokawinat", kalaYok)
	}
}

func TestOutfitColorAdvicePreservesLegacyKaliFirstFallback(t *testing.T) {
	// Legacy example: current day Monday, Friday born, age 44.
	// Sri collides with Kali twice, so it falls back to Montri: [6, 3, 4].
	advice := buildOutfitColorAdvice(time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC), "friday", 44)
	if advice == nil {
		t.Fatal("expected outfit colour advice")
	}
	wantAuspicious := []int{6, 3, 4}
	wantInauspicious := []int{1, 8, 7}
	for index, want := range wantAuspicious {
		if got := advice.Auspicious[index].Number; got != want {
			t.Fatalf("auspicious[%d].Number = %d, want %d", index, got, want)
		}
	}
	for index, want := range wantInauspicious {
		if got := advice.Inauspicious[index].Number; got != want {
			t.Fatalf("inauspicious[%d].Number = %d, want %d", index, got, want)
		}
	}
	if advice.Auspicious[0].RoleTH != "มนตรี" || advice.Auspicious[2].RoleTH != "มนตรี" {
		t.Fatalf("fallback roles = %#v, want Montri for first and third set", advice.Auspicious)
	}
	if colors := advice.Auspicious[0].Colors; len(colors) != 3 || colors[0].Hex != "#00FFFF" {
		t.Fatalf("auspicious colours = %#v, want the legacy number-6 palette", colors)
	}
	if got, want := advice.Auspicious[0].PlanetNameTH, "ดาวศุกร์"; got != want {
		t.Fatalf("auspicious[0].PlanetNameTH = %q, want %q", got, want)
	}
	if got, want := advice.Auspicious[0].RoleMeaningTH, "ผู้ใหญ่เมตตาและอุปถัมภ์"; got != want {
		t.Fatalf("auspicious[0].RoleMeaningTH = %q, want %q", got, want)
	}
}

func TestOutfitColorAdviceNeverReturnsAColorAsBothSupportAndCaution(t *testing.T) {
	for _, weekday := range []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday} {
		for _, birthDay := range []string{"sunday", "monday", "tuesday", "wednesday1", "wednesday2", "thursday", "friday", "saturday"} {
			for age := 1; age <= 99; age++ {
				advice := buildOutfitColorAdvice(time.Date(2026, time.August, 30+int(weekday), 0, 0, 0, 0, time.UTC), birthDay, age)
				if advice == nil {
					t.Fatalf("expected advice for %s, age %d", birthDay, age)
				}
				supporting := map[string]bool{}
				for _, set := range advice.Auspicious {
					for _, color := range set.Colors {
						supporting[color.Hex] = true
					}
				}
				for _, set := range advice.Inauspicious {
					for _, color := range set.Colors {
						if supporting[color.Hex] {
							t.Fatalf("%s (age %d) returns %s as both support and caution", birthDay, age, color.Hex)
						}
					}
				}
			}
		}
	}
}

func eqExpression(field string, value any) domain.RuleExpression {
	raw, _ := json.Marshal(value)
	return domain.RuleExpression{Field: field, Operator: "eq", Value: raw}
}

func TestEvaluateDayTaksa(t *testing.T) {
	// Natal: Sunday born -> Wednesday is Sri (+15), Friday is Kalakini (-35)
	taksaWed := EvaluateDayTaksa("sunday", 0, time.Wednesday)
	if taksaWed == nil || taksaWed.Name != "วันศรีประจำวันเกิด" || taksaWed.Score != 15 {
		t.Fatalf("Sunday -> Wednesday expected Sri +15, got %#v", taksaWed)
	}

	taksaFri := EvaluateDayTaksa("sunday", 0, time.Friday)
	if taksaFri == nil || taksaFri.Name != "วันกาลกิณีประจำวันเกิด" || taksaFri.Score != -35 {
		t.Fatalf("Sunday -> Friday expected Kalakini -35, got %#v", taksaFri)
	}

	// Progressed (ทักษาจร 9 จังหวะ):
	// Born Sunday, Age 2 (ตกตากลาง ๙): Saturday (7) is Kalakini (-35), Thursday (5) is Sri (+15)
	taksaCenterSat := EvaluateDayTaksa("sunday", 2, time.Saturday)
	if taksaCenterSat == nil || taksaCenterSat.Name != "วันกาลกิณีจร (อายุย่าง 2 ปี)" || taksaCenterSat.Score != -35 {
		t.Fatalf("Sunday Age 2 (Center) -> Saturday expected Kalakini Chor -35, got %#v", taksaCenterSat)
	}
	taksaCenterThu := EvaluateDayTaksa("sunday", 2, time.Thursday)
	if taksaCenterThu == nil || taksaCenterThu.Name != "วันศรีจร (อายุย่าง 2 ปี)" || taksaCenterThu.Score != 15 {
		t.Fatalf("Sunday Age 2 (Center) -> Thursday expected Sri Chor +15, got %#v", taksaCenterThu)
	}

	// Born Sunday, Age 3 (ก้าวออกจากตากลางไปตก 2 จันทร์): Sunday is Kalakini (-35), Saturday is Sri (+15)
	taksa3Sat := EvaluateDayTaksa("sunday", 3, time.Saturday)
	if taksa3Sat == nil || taksa3Sat.Name != "วันศรีจร (อายุย่าง 3 ปี)" || taksa3Sat.Score != 15 {
		t.Fatalf("Sunday Age 3 -> Saturday expected Sri Chor +15, got %#v", taksa3Sat)
	}
}

func TestEvaluateDayTaksaUsesEditableDatabaseTexts(t *testing.T) {
	texts := map[string]string{
		"taksa.role.kalakini.name":                  "ข้อควรระวัง",
		"taksa.role.kalakini.meaning":               "ข้อความความหมายที่แอดมินแก้ไข",
		"taksa.template.progressed_reason_name":     "{role}ประจำวัย {age}",
		"taksa.template.progressed_center_kalakini": "{date}: {role} สำหรับอายุ {age}",
	}

	// Sunday born, progressed age 2 lands in the center; Saturday is Kalakini.
	result := EvaluateDayTaksaWithTexts("sunday", 2, time.Saturday, texts)
	if result == nil {
		t.Fatal("expected Taksa result")
	}
	if result.Name != "ข้อควรระวังประจำวัย 2" {
		t.Fatalf("name = %q, want admin-edited template", result.Name)
	}
	if result.Meaning != "ข้อความความหมายที่แอดมินแก้ไข" {
		t.Fatalf("meaning = %q, want admin-edited meaning", result.Meaning)
	}
	if result.Explanation != "วันเสาร์: ข้อควรระวัง สำหรับอายุ 2" {
		t.Fatalf("explanation = %q, want admin-edited explanation", result.Explanation)
	}
}

func TestEvaluateAuspiciousDateWithDayOfBirth(t *testing.T) {
	// 2026-01-01 was Thursday
	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	lunar, err := thailunar.FromGregorian(date.Year(), int(date.Month()), date.Day())
	if err != nil {
		t.Fatal(err)
	}
	facts := factsForDate(date, lunar, false)
	rules := []domain.AuspiciousRule{}

	// Tuesday born: Thursday is Sri (+15)
	req := domain.AuspiciousSearchRequest{DayOfBirth: "tuesday"}
	result, blocked, err := evaluateAuspiciousDate(date, lunar, facts, rules, req)
	if err != nil {
		t.Fatal(err)
	}
	if blocked {
		t.Fatal("date should not be blocked")
	}
	if result.Score != 65 {
		t.Fatalf("expected score 65 (base 50 + Sri 15), got %d", result.Score)
	}
	foundSri := false
	for _, r := range result.Reasons {
		if r.Name == "วันศรีประจำวันเกิด" {
			foundSri = true
			break
		}
	}
	if !foundSri {
		t.Fatal("expected reason 'วันศรีประจำวันเกิด' to be present")
	}
}

type mockRuleRepo struct {
	rules []domain.AuspiciousRule
}

func (m *mockRuleRepo) ListActivities(ctx context.Context) ([]domain.AuspiciousActivity, error) {
	return []domain.AuspiciousActivity{{Code: "test", NameTH: "ทดสอบ"}}, nil
}

func (m *mockRuleRepo) ListRules(ctx context.Context, activityCode string) ([]domain.AuspiciousRule, error) {
	return m.rules, nil
}

func TestSearchAuspiciousDatesWithBirthDate(t *testing.T) {
	repo := &mockRuleRepo{
		rules: []domain.AuspiciousRule{
			{Code: "dummy", NameTH: "กฎทั่วไป", Effect: "score", Score: 5, Priority: 1, Condition: eqExpression("lunar_day", 1)},
		},
	}
	svc := NewAuspiciousService(repo)
	svc.clock = func() time.Time {
		return time.Date(2026, time.August, 28, 12, 0, 0, 0, time.FixedZone("Asia/Bangkok", 7*60*60))
	}
	// Born 10 July 1982 at 02:00: astrologically Friday, current progressed
	// age 45. The June search start is before this year's birthday and must not
	// lower the progressed age to 44.
	resp, err := svc.Search(context.Background(), domain.AuspiciousSearchRequest{
		ActivityCode: "test",
		StartDate:    "2026-06-01",
		EndDate:      "2026-06-07",
		BirthDate:    "1982-07-10",
		BirthTime:    "02:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
	foundProgressedAge := false
	for _, recommendation := range resp.Recommendations {
		for _, reason := range recommendation.Reasons {
			if strings.Contains(reason.Name, "อายุย่าง 45 ปี") {
				foundProgressedAge = true
			}
		}
	}
	if !foundProgressedAge {
		t.Fatalf("expected Taksa reasons to use current progressed age 45: %#v", resp.Recommendations)
	}
}

func TestSearchRejectsInvalidBirthDetails(t *testing.T) {
	repo := &mockRuleRepo{rules: []domain.AuspiciousRule{
		{Code: "dummy", NameTH: "กฎทั่วไป", Effect: "score", Score: 5, Priority: 1, Condition: eqExpression("lunar_day", 1)},
	}}
	svc := NewAuspiciousService(repo)
	tests := []struct {
		name string
		req  domain.AuspiciousSearchRequest
	}{
		{
			name: "malformed birth date",
			req:  domain.AuspiciousSearchRequest{ActivityCode: "test", StartDate: "2026-09-01", EndDate: "2026-09-07", BirthDate: "01/09/2000"},
		},
		{
			name: "future birth date",
			req:  domain.AuspiciousSearchRequest{ActivityCode: "test", StartDate: "2026-09-01", EndDate: "2026-09-07", BirthDate: "2027-01-01"},
		},
		{
			name: "malformed birth time",
			req:  domain.AuspiciousSearchRequest{ActivityCode: "test", StartDate: "2026-09-01", EndDate: "2026-09-07", BirthDate: "2000-01-01", BirthTime: "25:00"},
		},
		{
			name: "time without date",
			req:  domain.AuspiciousSearchRequest{ActivityCode: "test", StartDate: "2026-09-01", EndDate: "2026-09-07", BirthTime: "09:00"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.Search(context.Background(), tt.req); err == nil {
				t.Fatal("expected invalid birth details to be rejected")
			}
		})
	}
}

func TestDetermineAstrologicalBirthDay(t *testing.T) {
	// 2026-03-15 is Sunday
	sunDate := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)

	// Born Sunday 02:30 AM -> should shift to Saturday
	dayCode, dayThai := DetermineAstrologicalBirthDay(sunDate, "02:30")
	if dayCode != "saturday" || dayThai != "วันเสาร์" {
		t.Fatalf("expected Saturday for 02:30 AM, got %s (%s)", dayCode, dayThai)
	}

	// Born Sunday 08:30 AM -> should be Sunday
	dayCode, dayThai = DetermineAstrologicalBirthDay(sunDate, "08:30")
	if dayCode != "sunday" || dayThai != "วันอาทิตย์" {
		t.Fatalf("expected Sunday for 08:30 AM, got %s (%s)", dayCode, dayThai)
	}

	// 2026-03-18 is Wednesday
	wedDate := time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC)

	// Born Wednesday 10:00 AM -> Wednesday daytime
	dayCode, _ = DetermineAstrologicalBirthDay(wedDate, "10:00")
	if dayCode != "wednesday1" {
		t.Fatalf("expected wednesday1 for 10:00 AM, got %s", dayCode)
	}

	// Born Wednesday 20:00 PM -> Wednesday nighttime (Rahu)
	dayCode, _ = DetermineAstrologicalBirthDay(wedDate, "20:00")
	if dayCode != "wednesday2" {
		t.Fatalf("expected wednesday2 for 20:00 PM, got %s", dayCode)
	}

	// 2026-03-19 is Thursday
	thuDate := time.Date(2026, time.March, 19, 0, 0, 0, 0, time.UTC)

	// Born Thursday 04:00 AM -> shifts to Wednesday nighttime (Rahu)
	dayCode, _ = DetermineAstrologicalBirthDay(thuDate, "04:00")
	if dayCode != "wednesday2" {
		t.Fatalf("expected wednesday2 for Thursday 04:00 AM, got %s", dayCode)
	}
}

func TestWeddingRiangMonRuleEvaluation(t *testing.T) {
	rule := domain.AuspiciousRule{
		Code:        "thai_wedding_riang_mon",
		SignalGroup: "wedding_riang_mon",
		NameTH:      "วันเรียงหมอน (เคียงหมอน)",
		Effect:      "score",
		Score:       15,
		Priority:    450,
		Condition: domain.RuleExpression{
			Any: []domain.RuleExpression{
				{
					All: []domain.RuleExpression{
						{Field: "lunar_phase", Operator: "eq", Value: json.RawMessage(`"waxing"`)},
						{Field: "lunar_day", Operator: "in", Value: json.RawMessage(`[7,10,13]`)},
					},
				},
				{
					All: []domain.RuleExpression{
						{Field: "lunar_phase", Operator: "eq", Value: json.RawMessage(`"waning"`)},
						{Field: "lunar_day", Operator: "in", Value: json.RawMessage(`[4,8,10,14]`)},
					},
				},
			},
		},
	}

	tests := []struct {
		phase    string
		day      int
		expected bool
	}{
		// Waxing matches
		{"waxing", 7, true},
		{"waxing", 10, true},
		{"waxing", 13, true},
		// Waxing non-matches
		{"waxing", 1, false},
		{"waxing", 8, false},
		{"waxing", 14, false},
		// Waning matches
		{"waning", 4, true},
		{"waning", 8, true},
		{"waning", 10, true},
		{"waning", 14, true},
		// Waning non-matches
		{"waning", 1, false},
		{"waning", 7, false},
		{"waning", 13, false},
		{"waning", 15, false},
	}

	for _, tt := range tests {
		facts := auspiciousFacts{
			lunarPhase: tt.phase,
			lunarDay:   tt.day,
		}
		matched, err := evaluateExpression(rule.Condition, facts)
		if err != nil {
			t.Fatalf("unexpected error for %s %d: %v", tt.phase, tt.day, err)
		}
		if matched != tt.expected {
			t.Fatalf("%s %d: got matched=%v, want %v", tt.phase, tt.day, matched, tt.expected)
		}
	}
}

