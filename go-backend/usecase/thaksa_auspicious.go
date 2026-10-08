package usecase

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ThaksaPosition struct {
	Name        string
	Score       int
	Impact      string
	Explanation string
	Meaning     string
}

// Traditional Thai Taksa Wheel:
// อาทิตย์ (1) -> จันทร์ (2) -> อังคาร (3) -> พุธ (4) -> เสาร์ (7) -> พฤหัส (5) -> ราหู (8) -> ศุกร์ (6)
var taksaWheel = [...]int{1, 2, 3, 4, 7, 5, 8, 6}

func getStartPlanet(day string) int {
	switch strings.ToLower(strings.TrimSpace(day)) {
	case "sunday":
		return 1
	case "monday":
		return 2
	case "tuesday":
		return 3
	case "wednesday1", "wednesday":
		return 4
	case "saturday":
		return 7
	case "thursday":
		return 5
	case "wednesday2":
		return 8
	case "friday":
		return 6
	default:
		return 0
	}
}

func weekdayToPlanet(w time.Weekday) int {
	switch w {
	case time.Sunday:
		return 1
	case time.Monday:
		return 2
	case time.Tuesday:
		return 3
	case time.Wednesday:
		return 4
	case time.Thursday:
		return 5
	case time.Friday:
		return 6
	case time.Saturday:
		return 7
	default:
		return 1
	}
}

func planetIndexInWheel(planet int) int {
	for i, p := range taksaWheel {
		if p == planet {
			return i
		}
	}
	return 0
}

func getBirthDayThai(day string) string {
	switch strings.ToLower(strings.TrimSpace(day)) {
	case "sunday":
		return "วันอาทิตย์"
	case "monday":
		return "วันจันทร์"
	case "tuesday":
		return "วันอังคาร"
	case "wednesday1", "wednesday":
		return "วันพุธ (กลางวัน)"
	case "wednesday2":
		return "วันพุธ (กลางคืน)"
	case "thursday":
		return "วันพฤหัสบดี"
	case "friday":
		return "วันศุกร์"
	case "saturday":
		return "วันเสาร์"
	default:
		return ""
	}
}

// DetermineAstrologicalBirthDay determines the day of birth according to Thai astrology,
// where the astrological day starts at 06:00 AM (sunrise).
// Born 00:00 - 05:59 is shifted back to the previous day.
// Born Wednesday 18:00 - 05:59 is Wednesday nighttime (Rahu).
func DetermineAstrologicalBirthDay(birthDate time.Time, birthTimeStr string) (string, string) {
	hour := 9 // default daytime if unspecified
	if birthTimeStr != "" {
		parts := strings.Split(birthTimeStr, ":")
		if len(parts) >= 1 {
			var h int
			_, err := fmt.Sscanf(parts[0], "%d", &h)
			if err == nil {
				hour = h
			}
		}
	}

	astroDate := birthDate
	if hour < 6 {
		astroDate = astroDate.AddDate(0, 0, -1)
	}

	weekday := astroDate.Weekday()
	var dayCode string
	var dayThai string

	switch weekday {
	case time.Sunday:
		dayCode = "sunday"
		dayThai = "วันอาทิตย์"
	case time.Monday:
		dayCode = "monday"
		dayThai = "วันจันทร์"
	case time.Tuesday:
		dayCode = "tuesday"
		dayThai = "วันอังคาร"
	case time.Wednesday:
		// If originally Wednesday and hour >= 18, or originally Thursday before 06:00 (which shifted to Wednesday), it's Rahu (Wednesday Night).
		if (hour < 6) || (hour >= 18) {
			dayCode = "wednesday2"
			dayThai = "วันพุธ (กลางคืน/ราหู)"
		} else {
			dayCode = "wednesday1"
			dayThai = "วันพุธ (กลางวัน)"
		}
	case time.Thursday:
		dayCode = "thursday"
		dayThai = "วันพฤหัสบดี"
	case time.Friday:
		dayCode = "friday"
		dayThai = "วันศุกร์"
	case time.Saturday:
		dayCode = "saturday"
		dayThai = "วันเสาร์"
	}

	return dayCode, dayThai
}

// Standard 9-step Taksa Progression (เข้าตากลาง / ภูมิพระเกตุ ๙):
// 1 (อาทิตย์) -> 9 (ตากลาง) -> 2 (จันทร์) -> 3 (อังคาร) -> 4 (พุธ) -> 7 (เสาร์) -> 5 (พฤหัส) -> 8 (ราหู) -> 6 (ศุกร์)
var taksaProgression9 = [...]int{1, 9, 2, 3, 4, 7, 5, 8, 6}

type taksaRoleDefinition struct {
	code        string
	name        string
	score       int
	impact      string
	meaning     string
	description string
}

func taksaRoleDefinitions(texts map[string]string) [8]taksaRoleDefinition {
	defaults := [8]taksaRoleDefinition{
		{code: "boriwan", name: "บริวาร", score: 4, impact: "support", meaning: "ภูมิบริวารหมายถึงมิตรภาพ ผู้ช่วยเหลือ ผู้ร่วมงาน และความสามัคคี", description: "บริวาร มิตรภาพ ผู้ร่วมงาน ลูกน้อง และคนรอบข้าง"},
		{code: "ayu", name: "อายุ", score: 6, impact: "support", meaning: "ภูมิอายุหมายถึงสุขภาพ ความยั่งยืน ความราบรื่น และอายุยืนยาว", description: "อายุ สุขภาพ พลานามัย ความเป็นอยู่ และวิถีชีวิต"},
		{code: "dech", name: "เดช", score: 12, impact: "support", meaning: "ภูมิเดชหมายถึงอำนาจ บารมี ความเข้มแข็ง และความเป็นผู้นำ", description: "เดช อำนาจ บารมี เกียรติยศ และความก้าวหน้า"},
		{code: "sri", name: "ศรี", score: 15, impact: "support", meaning: "ภูมิศรีหมายถึงความดีงาม สิริมงคล โชคลาภ และความสำเร็จราบรื่น", description: "ศรี สิริมงคล โชคลาภ ทรัพย์สิน เมตตามหานิยม"},
		{code: "mula", name: "มูละ", score: 6, impact: "support", meaning: "ภูมิมูละหมายถึงความมั่นคง ทรัพย์สิน มรดก และหลักฐานมั่นคง", description: "มูละ ความมั่นคง หลักฐาน ทรัพย์สิน ที่อยู่อาศัย"},
		{code: "utsaha", name: "อุตสาหะ", score: 4, impact: "support", meaning: "ภูมิอุตสาหะหมายถึงความขยัน มุ่งมั่น มานะพยายามสู่ความสำเร็จ", description: "อุตสาหะ ความขยัน มุ่งมั่น ความเพียรพยายาม"},
		{code: "montri", name: "มนตรี", score: 10, impact: "support", meaning: "ภูมิมนตรีหมายถึงการอุปถัมภ์ค้ำชู ที่พึ่งพา และความเมตตาเอ็นดูจากผู้ใหญ่", description: "มนตรี ผู้ใหญ่อุปถัมภ์ เมตตาช่วยเหลือ ราบรื่น"},
		{code: "kalakini", name: "กาลกิณี", score: -35, impact: "caution", meaning: "กาลกิณีหมายถึงอุปสรรค สิ่งติดขัด หรือความไม่ราบรื่น โบราณมักให้หลีกเลี่ยง", description: "กาลกิณี อุปสรรค ความขัดแย้ง สิ่งที่ควรหลีกเลี่ยง"},
	}
	for i := range defaults {
		prefix := "taksa.role." + defaults[i].code + "."
		defaults[i].name = contentValue(texts, prefix+"name", defaults[i].name)
		defaults[i].meaning = contentValue(texts, prefix+"meaning", defaults[i].meaning)
		defaults[i].description = contentValue(texts, prefix+"description", defaults[i].description)
	}
	return defaults
}

func contentValue(texts map[string]string, key, fallback string) string {
	if value := strings.TrimSpace(texts[key]); value != "" {
		return value
	}
	return fallback
}

func renderAuspiciousTemplate(value string, replacements map[string]string) string {
	for key, replacement := range replacements {
		value = strings.ReplaceAll(value, "{"+key+"}", replacement)
	}
	return value
}

func planetIndexInProgression9(planet int) int {
	for i, p := range taksaProgression9 {
		if p == planet {
			return i
		}
	}
	return 0
}

// EvaluateDayTaksa evaluates either ทักษาจร (if age > 0) with 9-step center cell progression or ทักษากำเนิด (if age == 0).
func EvaluateDayTaksa(dayOfBirth string, age int, targetWeekday time.Weekday) *ThaksaPosition {
	return EvaluateDayTaksaWithTexts(dayOfBirth, age, targetWeekday, nil)
}

func EvaluateDayTaksaWithTexts(dayOfBirth string, age int, targetWeekday time.Weekday, texts map[string]string) *ThaksaPosition {
	startPlanet := getStartPlanet(dayOfBirth)
	if startPlanet == 0 {
		return nil
	}

	birthThai := getBirthDayThai(dayOfBirth)
	targetPlanet := weekdayToPlanet(targetWeekday)
	targetThai := [...]string{"วันอาทิตย์", "วันจันทร์", "วันอังคาร", "วันพุธ", "วันพฤหัสบดี", "วันศุกร์", "วันเสาร์"}[targetWeekday]

	roleDefs := taksaRoleDefinitions(texts)

	if age > 0 {
		// ทักษาจร: นับตามอายุย่าง 9 จังหวะ (อาทิตย์ -> ตากลาง -> จันทร์ -> อังคาร -> พุธ -> เสาร์ -> พฤหัส -> ราหู -> ศุกร์)
		startIdx9 := planetIndexInProgression9(startPlanet)
		step := (age - 1) % len(taksaProgression9)
		currIdx9 := (startIdx9 + step) % len(taksaProgression9)
		landed := taksaProgression9[currIdx9]

		var def taksaRoleDefinition
		isCenterCell := landed == 9

		if isCenterCell {
			// เมื่อตก "ตากลาง (๙)": ใช้ 7 (เสาร์) เป็นกาลี (กาลกิณีจร)
			// แมปภูมิ: 7=กาลกิณี, 5=ศรี, 1=เดช, 2=มนตรี, 4=มูละ, 6=อายุ, 8=อุตสาหะ, 3=บริวาร
			centerRoleMap := map[int]int{
				3: 0, // วันอังคาร = บริวาร
				6: 1, // วันศุกร์ = อายุ
				1: 2, // วันอาทิตย์ = เดช
				5: 3, // วันพฤหัสบดี = ศรี
				4: 4, // วันพุธ = มูละ
				8: 5, // วันราหู = อุตสาหะ
				2: 6, // วันจันทร์ = มนตรี
				7: 7, // วันเสาร์ = กาลกิณี (ตามตำราเมื่อตกตากลาง)
			}
			offset, ok := centerRoleMap[targetPlanet]
			if !ok {
				offset = 0
			}
			def = roleDefs[offset]
		} else {
			// ตกดาวเคราะห์ปกติ (1, 2, 3, 4, 7, 5, 8, 6): ดาวนั้นเป็นบริวารจร
			boriwanIdx := planetIndexInWheel(landed)
			targetIdx := planetIndexInWheel(targetPlanet)
			roleOffset := (targetIdx - boriwanIdx + len(taksaWheel)) % len(taksaWheel)
			def = roleDefs[roleOffset]
		}

		name := renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_reason_name", "วัน{role}จร (อายุย่าง {age} ปี)"), map[string]string{"role": def.name, "age": strconv.Itoa(age)})
		var explanation string
		replacements := map[string]string{"date": targetThai, "role": def.name, "birth_day": birthThai, "age": strconv.Itoa(age)}
		if def.code == "kalakini" {
			if isCenterCell {
				explanation = renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_center_kalakini", "{date} เป็นกาลกิณีจรสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี ตกภูมิกลาง) ตามหลักทักษาจร ควรหลีกเลี่ยงการเริ่มการมงคล"), replacements)
			} else {
				explanation = renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_kalakini", "{date} เป็นกาลกิณีจรประจำปีสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี) ตามหลักทักษาจร ควรหลีกเลี่ยงการเริ่มการมงคล"), replacements)
			}
		} else {
			explanation = renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_explanation", "{date} เป็นวัน{role}จรสำหรับผู้เกิด{birth_day} (อายุย่าง {age} ปี) ตามหลักทักษาจร"), replacements)
		}

		return &ThaksaPosition{
			Name:        name,
			Score:       def.score,
			Impact:      def.impact,
			Explanation: explanation,
			Meaning:     def.meaning,
		}
	}

	// ทักษากำเนิด (Natal Taksa)
	startIdx := planetIndexInWheel(startPlanet)
	targetIdx := planetIndexInWheel(targetPlanet)
	roleOffset := (targetIdx - startIdx + len(taksaWheel)) % len(taksaWheel)
	def := roleDefs[roleOffset]

	name := renderAuspiciousTemplate(contentValue(texts, "taksa.template.natal_reason_name", "วัน{role}ประจำวันเกิด"), map[string]string{"role": def.name})
	var explanation string
	if def.code == "kalakini" {
		explanation = renderAuspiciousTemplate(contentValue(texts, "taksa.template.natal_kalakini", "{date} เป็นกาลกิณีสำหรับผู้เกิด{birth_day} ตามหลักทักษาปกรณ์ ควรหลีกเลี่ยงการเริ่มการมงคล"), map[string]string{"date": targetThai, "birth_day": birthThai})
	} else {
		explanation = renderAuspiciousTemplate(contentValue(texts, "taksa.template.natal_explanation", "{date} เป็น{name}สำหรับผู้เกิด{birth_day} ตามหลักทักษาปกรณ์"), map[string]string{"date": targetThai, "name": name, "birth_day": birthThai})
	}

	return &ThaksaPosition{
		Name:        name,
		Score:       def.score,
		Impact:      def.impact,
		Explanation: explanation,
		Meaning:     def.meaning,
	}
}

type TaksaRoleInfo struct {
	RoleName           string `json:"role_name"`
	DayName            string `json:"day_name"`
	DayCode            string `json:"day_code"`
	PlanetNum          int    `json:"planet_num"`
	Symbol             string `json:"symbol"`
	Score              int    `json:"score"`
	Impact             string `json:"impact"`
	Description        string `json:"description"`
	Meaning            string `json:"meaning"`
	AuspiciousGuidance string `json:"auspicious_guidance"`
}

func taksaAuspiciousGuidance(def taksaRoleDefinition, dayName string) string {
	guidance := map[string]string{
		"boriwan":  "ช่วยเรื่องทีม ผู้ร่วมงาน และแรงสนับสนุน เหมาะกับกิจกรรมที่ต้องอาศัยความร่วมมือ",
		"ayu":      "ส่งเสริมความราบรื่น สุขภาพ และความยั่งยืน เหมาะกับสิ่งที่ต้องการความต่อเนื่อง",
		"dech":     "ส่งเสริมอำนาจ การตัดสินใจ และภาวะผู้นำ เหมาะกับการเริ่มงานสำคัญหรือการเจรจา",
		"sri":      "ส่งเสริมสิริมงคล โชคลาภ และเมตตามหานิยม เหมาะกับพิธีมงคลและการเริ่มต้นที่ต้องการความราบรื่น",
		"mula":     "ส่งเสริมความมั่นคง ทรัพย์สิน และรากฐาน เหมาะกับเรื่องบ้าน ที่ดิน และการวางหลักฐานระยะยาว",
		"utsaha":   "ส่งเสริมความเพียรและการลงมือทำ เหมาะกับงานที่ต้องใช้ความพยายามและความอดทน",
		"montri":   "ส่งเสริมการอุปถัมภ์และความช่วยเหลือจากผู้ใหญ่ เหมาะกับการขอคำปรึกษา การสนับสนุน และการเจรจา",
		"kalakini": "เป็นสัญญาณเตือนเรื่องอุปสรรคและความติดขัด ควรหลีกเลี่ยงการเริ่มพิธีหรือกิจกรรมสำคัญในวันนี้",
	}[def.code]
	return fmt.Sprintf("สำหรับคุณ %s อยู่ในภูมิ%s: %s", dayName, def.name, guidance)
}

type TaksaCalculationResult struct {
	IsTodayDefault       bool                     `json:"is_today_default"`
	BirthDayCode         string                   `json:"birth_day_code"`
	BirthDayName         string                   `json:"birth_day_name"`
	BirthPlanetNum       int                      `json:"birth_planet_num"`
	BirthTimeShifted     bool                     `json:"birth_time_shifted"`
	FullAge              int                      `json:"full_age"`
	ProgressedAge        int                      `json:"progressed_age"`
	IsCenterCell         bool                     `json:"is_center_cell"`
	LandedPlanetNum      int                      `json:"landed_planet_num"`
	LandedPlanetName     string                   `json:"landed_planet_name"`
	LandedPlanetSymbol   string                   `json:"landed_planet_symbol"`
	LandedPlanetCode     string                   `json:"landed_planet_code"`
	WheelRotationDegrees float64                  `json:"wheel_rotation_degrees"`
	HeroRoles            map[string]TaksaRoleInfo `json:"hero_roles"`
	AllRoles             []TaksaRoleInfo          `json:"all_roles"`
}

var PlanetMetadata = map[int]struct {
	Name   string
	Symbol string
	Code   string
}{
	1: {Name: "วันอาทิตย์", Symbol: "☀️", Code: "sunday"},
	2: {Name: "วันจันทร์", Symbol: "🌙", Code: "monday"},
	3: {Name: "วันอังคาร", Symbol: "⚔️", Code: "tuesday"},
	4: {Name: "วันพุธ (กลางวัน)", Symbol: "🌿", Code: "wednesday1"},
	7: {Name: "วันเสาร์", Symbol: "⚡", Code: "saturday"},
	5: {Name: "วันพฤหัสบดี", Symbol: "🎓", Code: "thursday"},
	8: {Name: "วันราหู/พุธคืน", Symbol: "🌌", Code: "wednesday2"},
	6: {Name: "วันศุกร์", Symbol: "💎", Code: "friday"},
}

// CalculateTaksa performs full 9-step Taksa calculation with center cell and today-default logic.
func CalculateTaksa(birthDateStr, birthTimeStr string, currentTime time.Time) TaksaCalculationResult {
	return CalculateTaksaWithTexts(birthDateStr, birthTimeStr, currentTime, nil)
}

func CalculateTaksaWithTexts(birthDateStr, birthTimeStr string, currentTime time.Time, texts map[string]string) TaksaCalculationResult {
	bangkok, _ := time.LoadLocation("Asia/Bangkok")
	if currentTime.IsZero() {
		currentTime = time.Now().In(bangkok)
	} else {
		currentTime = currentTime.In(bangkok)
	}

	roleDefs := taksaRoleDefinitions(texts)

	if strings.TrimSpace(birthDateStr) == "" {
		// Today default mode
		todayWeekday := currentTime.Weekday()
		todayPlanet := weekdayToPlanet(todayWeekday)
		meta := PlanetMetadata[todayPlanet]
		idxInWheel := planetIndexInWheel(todayPlanet)
		rotDeg := -float64(idxInWheel) * 45.0

		heroRoles := make(map[string]TaksaRoleInfo)
		allRoles := make([]TaksaRoleInfo, 0, 8)

		for i := 0; i < 8; i++ {
			pIdx := (idxInWheel + i) % 8
			pNum := taksaWheel[pIdx]
			pMeta := PlanetMetadata[pNum]

			info := TaksaRoleInfo{
				RoleName:           roleDefs[i].name,
				DayName:            fmt.Sprintf("%s (%d)", pMeta.Name, pNum),
				DayCode:            pMeta.Code,
				PlanetNum:          pNum,
				Symbol:             pMeta.Symbol,
				Score:              roleDefs[i].score,
				Impact:             roleDefs[i].impact,
				Description:        roleDefs[i].description,
				Meaning:            roleDefs[i].meaning,
				AuspiciousGuidance: taksaAuspiciousGuidance(roleDefs[i], pMeta.Name),
			}
			allRoles = append(allRoles, info)

			switch i {
			case 3:
				heroRoles["sri"] = info
			case 2:
				heroRoles["dech"] = info
			case 6:
				heroRoles["montri"] = info
			case 7:
				heroRoles["kalakini"] = info
			}
		}

		return TaksaCalculationResult{
			IsTodayDefault:       true,
			BirthDayCode:         meta.Code,
			BirthDayName:         meta.Name,
			BirthPlanetNum:       todayPlanet,
			BirthTimeShifted:     false,
			FullAge:              0,
			ProgressedAge:        0,
			IsCenterCell:         false,
			LandedPlanetNum:      todayPlanet,
			LandedPlanetName:     meta.Name,
			LandedPlanetSymbol:   meta.Symbol,
			LandedPlanetCode:     meta.Code,
			WheelRotationDegrees: rotDeg,
			HeroRoles:            heroRoles,
			AllRoles:             allRoles,
		}
	}

	bDate, err := time.ParseInLocation("2006-01-02", birthDateStr, bangkok)
	if err != nil {
		bDate, _ = time.Parse("2006-01-02", birthDateStr)
	}

	bDayCode, bDayThai := DetermineAstrologicalBirthDay(bDate, birthTimeStr)
	bStartPlanet := getStartPlanet(bDayCode)
	timeShifted := false
	if strings.TrimSpace(birthTimeStr) != "" {
		var h int
		fmt.Sscanf(strings.Split(birthTimeStr, ":")[0], "%d", &h)
		if h < 6 {
			timeShifted = true
		}
	}

	// Taksa age is the person's current age. It must not move when the user
	// changes the auspicious-date search range.
	fullAge, progressedAge := calculateCurrentAges(bDate, currentTime)

	// 9-Step Progression
	startIdx9 := planetIndexInProgression9(bStartPlanet)
	step := (progressedAge - 1) % len(taksaProgression9)
	currIdx9 := (startIdx9 + step) % len(taksaProgression9)
	landed := taksaProgression9[currIdx9]
	isCenterCell := landed == 9

	var rotDeg float64
	var landedMeta struct {
		Name   string
		Symbol string
		Code   string
	}

	if isCenterCell {
		landedMeta = struct {
			Name   string
			Symbol string
			Code   string
		}{Name: "ตากลาง (เกตุ)", Symbol: "✦", Code: "ketu"}
		rotDeg = 0.0
	} else {
		landedMeta = PlanetMetadata[landed]
		idxInWheel := planetIndexInWheel(landed)
		rotDeg = -float64(idxInWheel) * 45.0
	}

	heroRoles := make(map[string]TaksaRoleInfo)
	allRoles := make([]TaksaRoleInfo, 0, 8)

	if isCenterCell {
		// When landing in center cell 9:
		// 7=กาลกิณี, 5=ศรี, 1=เดช, 2=มนตรี, 4=มูละ, 6=อายุ, 8=อุตสาหะ, 3=บริวาร
		centerMapping := []struct {
			roleIdx   int
			planetNum int
		}{
			{0, 3}, // บริวาร = อังคาร (3)
			{1, 6}, // อายุ = ศุกร์ (6)
			{2, 1}, // เดช = อาทิตย์ (1)
			{3, 5}, // ศรี = พฤหัสบดี (5)
			{4, 4}, // มูละ = พุธ (4)
			{5, 8}, // อุตสาหะ = ราหู (8)
			{6, 2}, // มนตรี = จันทร์ (2)
			{7, 7}, // กาลกิณี = เสาร์ (7)
		}
		for _, m := range centerMapping {
			pMeta := PlanetMetadata[m.planetNum]
			info := TaksaRoleInfo{
				RoleName:           renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_role_name", "{role}จร"), map[string]string{"role": roleDefs[m.roleIdx].name}),
				DayName:            fmt.Sprintf("%s (%d)", pMeta.Name, m.planetNum),
				DayCode:            pMeta.Code,
				PlanetNum:          m.planetNum,
				Symbol:             pMeta.Symbol,
				Score:              roleDefs[m.roleIdx].score,
				Impact:             roleDefs[m.roleIdx].impact,
				Description:        roleDefs[m.roleIdx].description,
				Meaning:            roleDefs[m.roleIdx].meaning,
				AuspiciousGuidance: taksaAuspiciousGuidance(roleDefs[m.roleIdx], pMeta.Name),
			}
			allRoles = append(allRoles, info)
			switch m.roleIdx {
			case 3:
				heroRoles["sri"] = info
			case 2:
				heroRoles["dech"] = info
			case 6:
				heroRoles["montri"] = info
			case 7:
				heroRoles["kalakini"] = info
			}
		}
	} else {
		boriwanIdx := planetIndexInWheel(landed)
		for i := 0; i < 8; i++ {
			pIdx := (boriwanIdx + i) % 8
			pNum := taksaWheel[pIdx]
			pMeta := PlanetMetadata[pNum]
			info := TaksaRoleInfo{
				RoleName:           renderAuspiciousTemplate(contentValue(texts, "taksa.template.progressed_role_name", "{role}จร"), map[string]string{"role": roleDefs[i].name}),
				DayName:            fmt.Sprintf("%s (%d)", pMeta.Name, pNum),
				DayCode:            pMeta.Code,
				PlanetNum:          pNum,
				Symbol:             pMeta.Symbol,
				Score:              roleDefs[i].score,
				Impact:             roleDefs[i].impact,
				Description:        roleDefs[i].description,
				Meaning:            roleDefs[i].meaning,
				AuspiciousGuidance: taksaAuspiciousGuidance(roleDefs[i], pMeta.Name),
			}
			allRoles = append(allRoles, info)
			switch i {
			case 3:
				heroRoles["sri"] = info
			case 2:
				heroRoles["dech"] = info
			case 6:
				heroRoles["montri"] = info
			case 7:
				heroRoles["kalakini"] = info
			}
		}
	}

	return TaksaCalculationResult{
		IsTodayDefault:       false,
		BirthDayCode:         bDayCode,
		BirthDayName:         bDayThai,
		BirthPlanetNum:       bStartPlanet,
		BirthTimeShifted:     timeShifted,
		FullAge:              fullAge,
		ProgressedAge:        progressedAge,
		IsCenterCell:         isCenterCell,
		LandedPlanetNum:      landed,
		LandedPlanetName:     landedMeta.Name,
		LandedPlanetSymbol:   landedMeta.Symbol,
		LandedPlanetCode:     landedMeta.Code,
		WheelRotationDegrees: rotDeg,
		HeroRoles:            heroRoles,
		AllRoles:             allRoles,
	}
}

func calculateCurrentAges(birthDate, currentTime time.Time) (fullAge, progressedAge int) {
	fullAge = currentTime.Year() - birthDate.Year()
	if currentTime.Month() < birthDate.Month() || (currentTime.Month() == birthDate.Month() && currentTime.Day() < birthDate.Day()) {
		fullAge--
	}
	if fullAge < 0 {
		fullAge = 0
	}
	return fullAge, fullAge + 1
}
