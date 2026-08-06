package pipeline

import (
	"strings"
)

// Planet letter mappings in Thai Taksa
var planetChars = map[int]string{
	1: "ะาำิีึืุูเแโใไฤฦ็", // อาทิตย์ (1) - สระ
	2: "กขคฆง",            // จันทร์ (2)
	3: "จฉชซฌญ",           // อังคาร (3)
	4: "ฎฏฐฑฒณ",           // พุธ (4)
	7: "ดตถทธน",           // เสาร์ (7)
	5: "บปผฝพฟภม",          // พฤหัสบดี (5)
	8: "ยรลว",             // ราหู (8)
	6: "ศษสหฬฮ",           // ศุกร์ (6)
}

// Order of planets on the traditional Thai Taksa Wheel: อาทิตย์(1)->จันทร์(2)->อังคาร(3)->พุธ(4)->เสาร์(7)->พฤหัส(5)->ราหู(8)->ศุกร์(6)
var taksaWheel = []int{1, 2, 3, 4, 7, 5, 8, 6}

// GetTaksaCategoryRunes returns character map for a specific day and category (sri, dech, montri, mula, utsaha, ayu, bariwan)
func GetTaksaCategoryRunes(dayOfBirth string, category string) map[rune]bool {
	day := strings.ToLower(strings.TrimSpace(dayOfBirth))
	cat := strings.ToLower(strings.TrimSpace(category))
	if cat == "" {
		cat = "sri" // Default priority
	}

	startPlanet := getStartPlanet(day)
	if startPlanet == 0 {
		startPlanet = 1 // Default Sunday
	}

	// Thai Taksa Category sequence relative to บริวาร (Planet of Day of Birth):
	// บริวาร (+0), อายุ (+1), เดช (+2), ศรี (+3), มูละ (+4), อุตสาหะ (+5), มนตรี (+6), กาลกิณี (+7)
	// Example for Sunday (1):
	// - บริวาร (0) -> Planet 1 (สระ)
	// - อายุ (+1) -> Planet 2 (ก ข ค ฆ ง)
	// - เดช (+2) -> Planet 3 (จ ฉ ช ซ ฌ ญ)
	// - ศรี (+3) -> Planet 4 (ฎ ฏ ฐ ฑ ฒ ณ)
	// - มูละ (+4) -> Planet 7 (ด ต ถ ท ธ น)
	// - อุตสาหะ (+5) -> Planet 5 (บ ป ผ ฝ พ ฟ ภ ม)
	// - มนตรี (+6) -> Planet 8 (ย ร ล ว)
	// - กาลกิณี (+7) -> Planet 6 (ศ ษ ส ห ฬ ฮ)
	offsetMap := map[string]int{
		"bariwan": 0,
		"ayu":     1,
		"dech":    2,
		"sri":     3,
		"mula":    4,
		"utsaha":  5,
		"montri":  6,
	}

	offset, ok := offsetMap[cat]
	if !ok {
		offset = 3 // Default to Sri
	}

	startIdx := 0
	for i, p := range taksaWheel {
		if p == startPlanet {
			startIdx = i
			break
		}
	}

	targetPlanetIdx := (startIdx + offset) % len(taksaWheel)
	targetPlanet := taksaWheel[targetPlanetIdx]

	targetChars := planetChars[targetPlanet]
	res := make(map[rune]bool)
	for _, r := range targetChars {
		res[r] = true
	}
	return res
}

func getStartPlanet(day string) int {
	switch day {
	case "sunday":
		return 1
	case "monday":
		return 2
	case "tuesday":
		return 3
	case "wednesday1":
		return 4
	case "wednesday2":
		return 8
	case "thursday":
		return 5
	case "friday":
		return 6
	case "saturday":
		return 7
	default:
		return 1 // Default to Sunday wheel for general view when DayOfBirth is unselected
	}
}

// CountTaksaRunes counts how many characters of the chosen Taksa target exist in the name
func CountTaksaRunes(name string, dayOfBirth string, category string) int {
	targetMap := GetTaksaCategoryRunes(dayOfBirth, category)
	if len(targetMap) == 0 {
		return 0
	}

	count := 0
	for _, r := range name {
		if targetMap[r] {
			count++
		}
	}
	return count
}

// GetTaksaMatchedChars returns unique matched characters string (e.g. 'จ, ญ' or 'ก, ข')
func GetTaksaMatchedChars(name string, dayOfBirth string, category string) string {
	targetMap := GetTaksaCategoryRunes(dayOfBirth, category)
	if len(targetMap) == 0 {
		return "-"
	}

	seen := make(map[rune]bool)
	var matched []string
	for _, r := range name {
		if targetMap[r] && !seen[r] {
			seen[r] = true
			matched = append(matched, string(r))
		}
	}

	if len(matched) == 0 {
		return "-"
	}
	return strings.Join(matched, ", ")
}

// GetTaksaLabelThai returns human-readable Thai label for Taksa category
func GetTaksaLabelThai(category string) string {
	cat := strings.ToLower(strings.TrimSpace(category))
	switch cat {
	case "dech":
		return "เดช (อำนาจบารมี)"
	case "montri":
		return "มนตรี (ผู้ใหญ่อุปถัมภ์)"
	case "mula":
		return "มูละ (หลักทรัพย์/มรดก)"
	case "utsaha":
		return "อุตสาหะ (ความเพียร)"
	case "ayu":
		return "อายุ (สุขภาพ/อายุยืน)"
	case "bariwan":
		return "บริวาร (ครอบครัว/มิตรสหาย)"
	default:
		return "ศรี (มหาโชคลาภ/เงินทอง)"
	}
}
