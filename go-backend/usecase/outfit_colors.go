package usecase

import (
	"strings"
	"time"

	"go-backend/domain"
)

// This palette and ordering intentionally preserve the tested Nab Ha Sri
// clothing-colour logic from the legacy outfit-miracle project.
var outfitColorPalette = map[int][]domain.OutfitColor{
	1: {{NameTH: "สีแดง", Hex: "#FF0000"}, {NameTH: "สีแดงเลือดนก", Hex: "#CC0000"}, {NameTH: "แดงอ่อน", Hex: "#FF3333"}},
	2: {{NameTH: "สีขาว", Hex: "#FFFFFF"}, {NameTH: "สีเหลืองอ่อน", Hex: "#FFFFCC"}, {NameTH: "สีเหลืองอ่อน", Hex: "#FFFF99"}},
	3: {{NameTH: "สีชมพู", Hex: "#FF00FF"}, {NameTH: "สีชมพูอ่อน", Hex: "#FF66FF"}, {NameTH: "สีชมพูเข้ม", Hex: "#CC00CC"}},
	4: {{NameTH: "สีเขียวเข้ม", Hex: "#009900"}, {NameTH: "สีเขียวอ่อน", Hex: "#006600"}, {NameTH: "สีเขียวอ่อน", Hex: "#00CC00"}},
	5: {{NameTH: "สีเหลืองเข้ม", Hex: "#FFFF00"}, {NameTH: "สีแสด", Hex: "#FF8000"}, {NameTH: "สีส้มเข้ม", Hex: "#FFFF33"}},
	6: {{NameTH: "สีฟ้า", Hex: "#00FFFF"}, {NameTH: "สีฟ้าอ่อน", Hex: "#0080FF"}, {NameTH: "สีน้ำเงิน", Hex: "#0000FF"}},
	7: {{NameTH: "สีดำ", Hex: "#000000"}, {NameTH: "สีม่วง", Hex: "#6600CC"}, {NameTH: "สีม่วง", Hex: "#9933FF"}},
	8: {{NameTH: "สีน้ำตาล", Hex: "#663300"}, {NameTH: "สีเทา", Hex: "#808080"}, {NameTH: "สีเทาอ่อน", Hex: "#C0C0C0"}},
}

var outfitDaySequence = []int{1, 6, 8, 5, 7, 4, 3, 2}
var outfitAgeSequence = []int{6, 8, 5, 7, 4, 3, 2, 1}
var outfitForwardSequence = []int{1, 7, 2, 3, 4, 7, 5, 8, 6}

var outfitPlanetNames = map[int]string{
	1: "ดาวอาทิตย์", 2: "ดาวจันทร์", 3: "ดาวอังคาร", 4: "ดาวพุธ",
	5: "ดาวพฤหัสบดี", 6: "ดาวศุกร์", 7: "ดาวเสาร์", 8: "ดาวราหู",
}

var outfitRoleMeanings = map[string]string{
	"ศรี":   "สิริมงคล โชคลาภ และเสน่ห์",
	"มนตรี": "ผู้ใหญ่เมตตาและอุปถัมภ์",
	"กาลี":  "อุปสรรค จึงควรหลีกเลี่ยง",
}

type outfitSriKali struct {
	sri  int
	kali int
}

func buildOutfitColorAdvice(date time.Time, birthDay string, progressedAge int) *domain.OutfitColorAdvice {
	birthNumber := getStartPlanet(birthDay)
	if birthNumber == 0 || progressedAge < 1 {
		return nil
	}

	appointment := outfitSriKaliForDay(weekdayToPlanet(date.Weekday()))
	birth := outfitSriKaliForDay(birthNumber)
	age := outfitSriKaliForAge(birthNumber, progressedAge)

	contexts := []struct {
		context  string
		label    string
		result   outfitSriKali
		sequence []int
	}{
		{"appointment", "สีเสริมวันนัด", appointment, outfitDaySequence},
		{"birth_day", "สีเสริมวันเกิด", birth, outfitDaySequence},
		{"progressed_age", "สีเสริมอายุย่าง", age, outfitAgeSequence},
	}

	kaliSet := map[int]bool{}
	for _, item := range contexts {
		kaliSet[item.result.kali] = true
	}

	advice := &domain.OutfitColorAdvice{
		Auspicious:   make([]domain.OutfitColorSet, 0, len(contexts)),
		Inauspicious: make([]domain.OutfitColorSet, 0, len(contexts)),
	}
	seenAuspicious := map[int]bool{}
	for _, item := range contexts {
		advice.Inauspicious = append(advice.Inauspicious, outfitColorSet(item.context, strings.Replace(item.label, "สีเสริม", "สีควรเลี่ยง", 1), "กาลี", item.result.kali))

		number, role := item.result.sri, "ศรี"
		if kaliSet[number] || seenAuspicious[number] {
			number, role = outfitMontriNumber(number, item.sequence), "มนตรี"
		}
		// กาลี keeps priority. Do not invent another role if มนตรี also
		// conflicts; omitting this supporting set is clearer than giving the
		// user a contradictory colour recommendation.
		if kaliSet[number] || seenAuspicious[number] {
			continue
		}
		seenAuspicious[number] = true
		advice.Auspicious = append(advice.Auspicious, outfitColorSet(item.context, item.label, role, number))
	}
	return advice
}

func outfitSriKaliForDay(start int) outfitSriKali {
	return outfitSriKali{
		sri:  outfitMove(start, 5, outfitDaySequence),
		kali: outfitMove(start, 1, outfitDaySequence),
	}
}

func outfitSriKaliForAge(birthNumber, age int) outfitSriKali {
	finalNumber := outfitMove(birthNumber, outfitSingleDigit(age)-1, outfitForwardSequence)
	return outfitSriKali{
		sri:  outfitMove(finalNumber, 5, outfitAgeSequence),
		kali: outfitMove(finalNumber, 1, outfitAgeSequence),
	}
}

func outfitSingleDigit(value int) int {
	for value >= 10 {
		value = value/10 + value%10
	}
	return value
}

func outfitMove(start, moves int, sequence []int) int {
	for index, number := range sequence {
		if number == start {
			return sequence[(index+moves)%len(sequence)]
		}
	}
	return start
}

func outfitMontriNumber(sriNumber int, sequence []int) int {
	for index, number := range sequence {
		if number == sriNumber {
			return sequence[(index-3+len(sequence))%len(sequence)]
		}
	}
	return sriNumber
}

func outfitColorSet(context, label, role string, number int) domain.OutfitColorSet {
	colors := append([]domain.OutfitColor(nil), outfitColorPalette[number]...)
	return domain.OutfitColorSet{
		Context:       context,
		LabelTH:       label,
		RoleTH:        role,
		RoleMeaningTH: outfitRoleMeanings[role],
		PlanetNameTH:  outfitPlanetNames[number],
		Number:        number,
		Colors:        colors,
	}
}
