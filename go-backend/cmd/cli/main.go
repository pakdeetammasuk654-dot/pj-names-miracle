package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-backend/config"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
	"go-backend/usecase/pipeline"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BgBlue    = "\033[44m"
	BgGreen   = "\033[42m"
	BgMagenta = "\033[45m"
	BgYellow  = "\033[43m"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	printHeader()

	fmt.Printf("%sConnecting to Miracle Names Database...%s\n", Yellow, Reset)
	db, err := config.InitDB()
	if err != nil {
		fmt.Printf("%s[Error] Database connection failed: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewNameRepository(db)
	useCase := usecase.NewNameSearchUseCase(repo)

	for {
		fmt.Println()
		fmt.Printf("%s===========================================================%s\n", Cyan, Reset)
		fmt.Printf("%s                🔍 ค้นหาชื่อมงคล (TOP 100)                %s\n", Bold+Yellow, Reset)
		fmt.Printf("%s===========================================================%s\n", Cyan, Reset)

		fmt.Printf("\n%sกรุณาเลือกโหมดการค้นหา:%s\n", Bold+White, Reset)
		fmt.Printf("  %s1)%s รวมชื่อ + นามสกุล (Combined Mode)\n", Green, Reset)
		fmt.Printf("  %s2)%s เฉพาะชื่ออย่างเดียว (Standalone Mode)\n", Green, Reset)
		fmt.Printf("  %s3)%s 🤖 ค้นหาด้วย AI (Semantic Wish Search / คำอธิษฐานธรรมชาติ)\n", Yellow, Reset)
		fmt.Printf("  %s4)%s 📖 อ่านคู่มือกระบวนการคิดชื่อมงคล & หลักเกณฑ์จัดอันดับ (Methodology Guide)\n", Cyan, Reset)
		fmt.Printf("  %s0)%s ออกจากโปรแกรม (Exit)\n", Red, Reset)
		fmt.Printf("%sเลือกข้อ (0-4): %s", Bold+Cyan, Reset)

		modeChoice := readInput(reader)
		if modeChoice == "0" {
			fmt.Printf("\n%sขอบคุณที่ใช้งาน Miracle Names CLI! สวัสดีครับ 👋%s\n\n", Green, Reset)
			break
		}

		if modeChoice == "4" {
			renderMethodologyGuide(reader)
			continue
		}

		var mode domain.SearchMode
		var surname string
		var wishQuery string

		if modeChoice == "3" {
			mode = domain.ModeSemantic
			fmt.Printf("\n%sเลือกคำอธิษฐานมงคลลัด (Quick Preset Wish):%s\n", Bold+White, Reset)
			fmt.Printf("  %s1)%s 🎓 ปัญญาเฉลียวฉลาด เรียนเก่ง ประสบความสำเร็จทางการศึกษา\n", Yellow, Reset)
			fmt.Printf("  %s2)%s 💰 ร่ำรวยเป็นเศรษฐี มีทรัพย์สินมั่งคั่ง ค้าขายเจริญรุ่งเรือง\n", Yellow, Reset)
			fmt.Printf("  %s3)%s 👑 ผู้นำมีอำนาจบารมี ยศถาบรรดาศักดิ์ เป็นที่เคารพนับถือ\n", Yellow, Reset)
			fmt.Printf("  %s4)%s 💖 เสน่ห์เมตตามหานิยม ผู้ใหญ่อุปถัมภ์คอยช่วยเหลือ\n", Yellow, Reset)
			fmt.Printf("  %s5)%s 🩺 ร่างกายแข็งแรง ปราศจากโรคภัย มีอายุยืนยาวพูนสุข\n", Yellow, Reset)
			fmt.Printf("  %s6)%s ✍️  พิมพ์คำอธิษฐานเอง (Custom Wish Text)\n", Cyan, Reset)
			fmt.Printf("%sเลือกข้อ (1-6) [Default 1]: %s", Bold+Cyan, Reset)

			presetChoice := readInput(reader)
			wishQuery = mapPresetWish(presetChoice, reader)
		} else {
			if modeChoice == "1" {
				mode = domain.ModeCombined
				fmt.Printf("\n%sกรอกนามสกุลของคุณ (เช่น สมบูรณ์, ทองแสน): %s", Bold+Yellow, Reset)
				surname = readInput(reader)
				if surname == "" {
					surname = "สมบูรณ์"
				}
			} else {
				mode = domain.ModeStandalone
			}

			fmt.Printf("\n%sต้องการเน้นความหมายคำอธิษฐานด้วย AI ร่วมด้วยหรือไม่? (AI Semantic Wish Focus):%s\n", Bold+White, Reset)
			fmt.Printf("  %s0)%s 🌸 เน้นสิริมงคลพื้นดวง (นำพาโชคดีและความสำเร็จครอบคลุมทุกมิติชีวิต)\n", Yellow, Reset)
			fmt.Printf("  %s1)%s 🎓 ปัญญาเฉลียวฉลาด เรียนเก่ง\n", Yellow, Reset)
			fmt.Printf("  %s2)%s 💰 ร่ำรวยเป็นเศรษฐี ทรัพย์สินมั่งคั่ง\n", Yellow, Reset)
			fmt.Printf("  %s3)%s 👑 ผู้นำมีอำนาจบารมี ยศถาบรรดาศักดิ์\n", Yellow, Reset)
			fmt.Printf("  %s4)%s 💖 เสน่ห์เมตตามหานิยม ผู้ใหญ่อุปถัมภ์\n", Yellow, Reset)
			fmt.Printf("  %s5)%s 🩺 สุขภาพแข็งแรง ปราศจากโรคภัย\n", Yellow, Reset)
			fmt.Printf("  %s6)%s ✍️  พิมพ์คำอธิษฐานเอง (Custom Wish Text)\n", Cyan, Reset)
			fmt.Printf("%sเลือกข้อ (0-6) [Default 0]: %s", Bold+Cyan, Reset)

			wishChoice := readInput(reader)
			if wishChoice != "" && wishChoice != "0" {
				wishQuery = mapPresetWish(wishChoice, reader)
			}
		}

		fmt.Printf("\n%sเลือกวันเกิด (คัดกาลกิณีออก 100%%):%s\n", Bold+White, Reset)
		fmt.Println("  0) ไม่ระบุวันเกิด")
		fmt.Println("  1) วันอาทิตย์")
		fmt.Println("  2) วันจันทร์")
		fmt.Println("  3) วันอังคาร")
		fmt.Println("  4) วันพุธ (กลางวัน)")
		fmt.Println("  5) วันพุธ (กลางคืน / ราหู)")
		fmt.Println("  6) วันพฤหัสบดี")
		fmt.Println("  7) วันศุกร์")
		fmt.Println("  8) วันเสาร์")
		fmt.Printf("%sเลือกข้อ (0-8): %s", Bold+Cyan, Reset)

		dayChoice := readInput(reader)
		dayOfBirth := mapDayChoice(dayChoice)

		taksaTarget := "sri"
		fmt.Printf("\n%sเลือกเป้าหมายมงคลที่ต้องการเน้นเป็นอันดับ #1 (Taksa Focus Priority):%s\n", Bold+White, Reset)
		fmt.Printf("  %s1)%s 🌟 ศรี (มหาโชคลาภ / เงินทอง / สิริมงคล) [แนะนำ]\n", Yellow, Reset)
		fmt.Printf("  %s2)%s 👑 เดช (อำนาจบารมี / ยศถาบรรดาศักดิ์ / ความเกรงขาม)\n", Yellow, Reset)
		fmt.Printf("  %s3)%s 🤝 มนตรี (ผู้ใหญ่อุปถัมภ์ / คนคอยช่วยเหลือดันงาน)\n", Yellow, Reset)
		fmt.Printf("  %s4)%s 🏠 มูละ (หลักทรัพย์ / มรดก / ฐานะมั่นคง)\n", Yellow, Reset)
		fmt.Printf("  %s5)%s ⚡️ อุตสาหะ (ความเพียร / มุ่งมั่นฟันฝ่าสำเร็จ)\n", Yellow, Reset)
		fmt.Printf("  %s6)%s 🩺 อายุ (สุขภาพแข็งแรง / ปราศจากโรคภัย)\n", Yellow, Reset)
		fmt.Printf("  %s7)%s 👥 บริวาร (มิตรสหาย / ครอบครัว / ลูกน้องซื่อสัตย์)\n", Yellow, Reset)
		fmt.Printf("%sเลือกเป้าหมาย (1-7) [Default 1-ศรี]: %s", Bold+Cyan, Reset)

		taksaChoice := readInput(reader)
		taksaTarget = mapTaksaChoice(taksaChoice)

		fmt.Printf("\n%sความยาวพยางค์ชื่อที่ต้องการ (Syllable Length Preference):%s\n", Bold+White, Reset)
		fmt.Printf("  %s1)%s 🟢 ทุกพยางค์ (รวม 1, 2, 3 พยางค์)\n", Green, Reset)
		fmt.Printf("  %s2)%s 🌟 เฉพาะชื่อ 2 พยางค์ขึ้นไป (กรองชื่อ 1 พยางค์ออก)\n", Yellow, Reset)
		fmt.Printf("%sเลือกข้อ (1-2) [Default 1]: %s", Bold+Cyan, Reset)
		lengthChoice := readInput(reader)
		minLength := 0
		if lengthChoice == "2" {
			minLength = 4
		}

		fmt.Printf("\n%sกำลังประมวลผลผ่าน Smart Auto-Ranking Pipeline (Mode: %s, Day: %s)...%s\n", Yellow, mode, dayOfBirth, Reset)

		req := domain.NameSearchRequest{
			Mode:        mode,
			Surname:     surname,
			DayOfBirth:  dayOfBirth,
			TaksaTarget: taksaTarget,
			WishQuery:   wishQuery,
			Limit:       100,
			MinLength:   minLength,
		}

		results, err := useCase.SearchTopNames(context.Background(), req)
		if err != nil {
			fmt.Printf("%s[Error] การค้นหาล้มเหลว: %v%s\n", Red, err, Reset)
			continue
		}

		displayResultsTable(results, mode, surname, dayOfBirth, taksaTarget, wishQuery, reader)
	}
}

func printHeader() {
	fmt.Printf("\033[H\033[2J")
	fmt.Printf("%s%s", BgBlue+White+Bold, strings.Repeat(" ", 88))
	fmt.Printf("\n  ✨  MIRACLE NAMES CLI - TERMINAL UI (3-TIER ULTIMATE HYBRID RANKING)  ✨  \n")
	fmt.Printf("%s%s\n", strings.Repeat(" ", 88), Reset)
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func mapDayChoice(choice string) string {
	switch choice {
	case "1":
		return "sunday"
	case "2":
		return "monday"
	case "3":
		return "tuesday"
	case "4":
		return "wednesday1"
	case "5":
		return "wednesday2"
	case "6":
		return "thursday"
	case "7":
		return "friday"
	case "8":
		return "saturday"
	default:
		return ""
	}
}

func mapTaksaChoice(choice string) string {
	switch choice {
	case "2":
		return "dech"
	case "3":
		return "montri"
	case "4":
		return "mula"
	case "5":
		return "utsaha"
	case "6":
		return "ayu"
	case "7":
		return "bariwan"
	default:
		return "sri"
	}
}

func mapPresetWish(choice string, reader *bufio.Reader) string {
	switch choice {
	case "2":
		return "ขอชื่อที่สื่อถึงความร่ำรวยเป็นเศรษฐี มีทรัพย์สินเงินทองมั่งคั่ง ค้าขายร่ำรวย และมีโชคลาภเข้ามาไม่ขาดสาย"
	case "3":
		return "อยากได้ชื่อที่สื่อถึงผู้นำที่มีอำนาจบารมี มีตำแหน่งยศถาบรรดาศักดิ์สูงส่ง มีความเด็ดขาด และเป็นที่เคารพนับถือ"
	case "4":
		return "ขอชื่อที่สื่อถึงเสน่ห์ อ่อนโยน มีผู้ใหญ่อุปถัมภ์คอยช่วยเหลือดันงาน และเป็นที่รักเอ็นดูของทุกคนที่พบเห็น"
	case "5":
		return "อยากได้ชื่อที่สื่อถึงความมีสุขภาพร่างกายแข็งแรง ปราศจากโรคภัยไข้เจ็บ มีจิตใจผ่องใส และมีชีวิตที่สมบูรณ์ราบรื่น"
	case "6":
		fmt.Printf("\n%sพิมพ์คำอธิษฐานของคุณเอง: %s", Bold+Yellow, Reset)
		custom := readInput(reader)
		if custom == "" {
			return "อยากได้ชื่อที่สื่อถึงปัญญาเฉลียวฉลาด เรียนเก่ง มีความคิดสร้างสรรค์ และประสบความสำเร็จในชีวิต"
		}
		return custom
	default:
		return "อยากได้ชื่อที่สื่อถึงปัญญาเฉลียวฉลาด เรียนเก่ง มีความคิดสร้างสรรค์ และประสบความสำเร็จในชีวิต"
	}
}

func displayResultsTable(results []domain.NameResultItem, mode domain.SearchMode, surname string, dayOfBirth string, taksaTarget string, wishQuery string, reader *bufio.Reader) {
	total := len(results)
	if total == 0 {
		fmt.Printf("\n%s[!] ไม่พบรายชื่อที่ผ่านเกณฑ์กาลกิณีและเลขมงคล%s\n", Red, Reset)
		return
	}

	pageSize := 15
	currentPage := 0
	totalPages := (total + pageSize - 1) / pageSize

	for {
		fmt.Printf("\033[H\033[2J")
		fmt.Printf("%s================================================================================================-------------------------%s\n", Cyan, Reset)
		if wishQuery != "" {
			fmt.Printf(" 🏆 3-TIER HYBRID RANKING (เน้น %s #1 -> AI Vector Match #2 -> คะแนนสุทธิ #3) (หน้า %d/%d)\n", pipeline.GetTaksaLabelThai(taksaTarget), currentPage+1, totalPages)
			fmt.Printf(" 💬 คำอธิษฐาน AI: %s%s%s\n", Yellow, truncateString(wishQuery, 85), Reset)
		} else {
			fmt.Printf(" 🏆 ผลการจัดอันดับ 100 ชื่อมงคล (เน้น %s #1 -> คะแนนสุทธิ #2) (หน้า %d/%d)\n", pipeline.GetTaksaLabelThai(taksaTarget), currentPage+1, totalPages)
		}

		if mode == domain.ModeCombined {
			fmt.Printf(" 📌 โหมด: รวมชื่อ + นามสกุล (%s%s%s) | วันเกิด: %s%s%s\n", Yellow, surname, Reset, Magenta, formatDayThai(dayOfBirth), Reset)
		} else if mode == domain.ModeStandalone {
			fmt.Printf(" 📌 โหมด: เฉพาะชื่ออย่างเดียว | วันเกิด: %s%s%s\n", Magenta, formatDayThai(dayOfBirth), Reset)
		} else {
			fmt.Printf(" 📌 โหมด: AI Semantic Search | วันเกิด: %s%s%s\n", Magenta, formatDayThai(dayOfBirth), Reset)
		}
		fmt.Printf("%s================================================================================================-------------------------%s\n", Cyan, Reset)

		fmt.Printf("%s%-6s %-14s %-12s %-8s %-12s %-12s %-11s %-11s %-25s%s\n",
			Bold+White, "อันดับ", "ชื่อมงคล", "ทักษาเน้น(#1)", "คะแนน", "เลขคู่ Sat", "เลขคู่ Sha", "เกรด Sat", "เกรด Sha", "ความหมายสรุปพลังตัวเลข", Reset)
		fmt.Println(strings.Repeat("-", 122))

		start := currentPage * pageSize
		end := start + pageSize
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			item := results[i]

			color := Green
			if item.Rank <= 3 {
				color = Yellow + Bold
			}

			satBadge := formatBadge(item.SatPairType)
			shaBadge := formatBadge(item.ShaPairType)

			combinedDesc := item.SatMiracleDesc
			if item.ShaMiracleDesc != "" {
				combinedDesc += " | " + item.ShaMiracleDesc
			}

			shortLabel := getShortTaksaLabel(taksaTarget)
			taksaDisplay := " - "
			if item.TaksaCount > 0 {
				taksaDisplay = fmt.Sprintf("🌟%s(%d)", shortLabel, item.TaksaCount)
			}

			perfectBadge := ""
			if item.IsPerfectCombined {
				perfectBadge = Yellow + " 👑[มงคลสมบูรณ์แบบ]" + Reset
			}

			fmt.Printf("%s#%-5d%s %-14s %-12s %s%-8d%s %-12s %-12s %-11s %-11s %-25s%s\n",
				color, item.Rank, Reset,
				Bold+item.THName+Reset,
				Yellow+taksaDisplay+Reset,
				Bold+Yellow, item.TotalScore, Reset,
				item.SatPairsDisplay,
				item.ShaPairsDisplay,
				satBadge,
				shaBadge,
				truncateString(combinedDesc, 20),
				perfectBadge,
			)
		}

		fmt.Printf("%s================================================================================================-------------------------%s\n", Cyan, Reset)
		fmt.Printf("%s[N] หน้าถัดไป | [P] หน้าถอยหลัง | [1-100] ดูการ์ดมงคลอวยพร | [Q] ออกเมนู | กรอกคำสั่ง: %s", Bold+Yellow, Reset)

		nav := strings.ToUpper(readInput(reader))
		if nav == "N" || nav == "NEXT" {
			if currentPage < totalPages-1 {
				currentPage++
			}
		} else if nav == "P" || nav == "PREV" {
			if currentPage > 0 {
				currentPage--
			}
		} else if nav == "Q" || nav == "QUIT" || nav == "" {
			break
		} else {
			rankNum, err := strconv.Atoi(nav)
			if err == nil && rankNum >= 1 && rankNum <= total {
				renderBlessingCard(results[rankNum-1], surname, dayOfBirth, taksaTarget, wishQuery, reader)
			}
		}
	}
}

func renderBlessingCard(item domain.NameResultItem, surname string, dayOfBirth string, taksaTarget string, wishQuery string, reader *bufio.Reader) {
	fmt.Printf("\033[H\033[2J")

	fullName := item.THName
	if surname != "" {
		fullName = item.THName + "  " + surname
	}

	meaning := item.Meaning
	if meaning == "" {
		meaning = "ผู้มีความเจริญรุ่งเรือง มั่งคั่ง และมีโชคลาภอันประเสริฐ"
	}

	if wishQuery == "" {
		wishQuery = "ขอให้ชีวิตเปี่ยมด้วยปัญญาเฉลียวฉลาด สุขภาพแข็งแรง และมั่งคั่งพูนสุข"
	}

	shortLabel := getShortTaksaLabel(taksaTarget)

	dayDisplay := formatDayThai(dayOfBirth)
	taksaLine := fmt.Sprintf("🌟 อักษร%s %d ตัว [%s]", shortLabel, item.TaksaCount, pipeline.GetTaksaLabelThai(taksaTarget))
	if dayOfBirth == "" {
		dayDisplay = "ไม่ระบุวันเกิด (ใช้ได้กับทุกวันเกิด)"
		taksaLine = "ไม่ระบุวันเกิด (เน้นพลังเลขศาสตร์เป็นหลัก)"
	} else {
		dayDisplay += "  " + fmt.Sprintf("%s[🛡️ ปลอดกาลกิณี 100%%]%s", Green+Bold, Reset)
	}

	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────────────────────┐%s\n", Cyan, Reset)
	fmt.Printf("%s│  ✨  MIRACLE NAME BLESSING CARD - การ์ดวิเคราะห์สิริมงคลประจำนาม  ✨                     │%s\n", Bold+BgMagenta+White, Reset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────────────────────┤%s\n", Cyan, Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│   👤 %sชื่อมงคลประจำตัว:%s  %s%s%s\n", Bold+White, Reset, Bold+Yellow, fullName, Reset)
	fmt.Printf("│   📖 %sความหมายอันลึกซึ้ง:%s  %s%s%s\n", Bold+White, Reset, Cyan, meaning, Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│   📅 %sวันเกิดประจำผัง:%s    %s%s%s\n", Bold+White, Reset, Magenta, dayDisplay, Reset)
	fmt.Printf("│   🌟 %sผังอักขระทักษา:%s    %s%s%s\n", Bold+White, Reset, Yellow, taksaLine, Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│   🔮 %sพลังเกรดตัวเลขมงคล:%s                                                           │\n", Bold+White, Reset)
	fmt.Printf("│      • เฉพาะชื่อ (Sat Sum %d):     %s [%s]%s ➔ %s\n", item.SatSum, BgGreen+White, item.SatPairType, Reset, truncateString(item.SatMiracleDesc, 32))
	fmt.Printf("│      • รวมนามสกุล (Sha Sum %d):   %s [%s]%s ➔ %s\n", item.ShaSum, BgGreen+White, item.ShaPairType, Reset, truncateString(item.ShaMiracleDesc, 32))
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│   💬 %sคำอธิษฐานมงคลประจำใจ:%s                                                          │\n", Bold+White, Reset)
	fmt.Printf("│      \"%s%s%s\"\n", Yellow, truncateString(wishQuery, 75), Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│   📜 %sคำอวยพรมงคลชีวิตประจำนาม:%s                                                      │\n", Bold+White, Reset)
	fmt.Printf("│      \"ขอให้นามมงคล '%s%s%s' นำพาชีวิตให้เปี่ยมด้วยโชคลาภ สติปัญญา                       │\n", Bold+Yellow, item.THName, Reset)
	fmt.Printf("│       และมีความเจริญรุ่งเรือง ก้าวหน้าประสบความสำเร็จในชีวิตสืบไป\"                      │\n")
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)
	fmt.Printf("\n%s[กด ENTER เพื่อกลับไปตารางจัดอันดับ]: %s", Bold+Yellow, Reset)

	readInput(reader)
}

func getShortTaksaLabel(category string) string {
	switch strings.ToLower(category) {
	case "dech":
		return "เดช"
	case "montri":
		return "มนตรี"
	case "mula":
		return "มูละ"
	case "utsaha":
		return "อุตสาหะ"
	case "ayu":
		return "อายุ"
	case "bariwan":
		return "บริวาร"
	default:
		return "ศรี"
	}
}

func formatBadge(pairType string) string {
	if pairType == "" {
		return " [-] "
	}
	if strings.Contains(pairType, "D") {
		return fmt.Sprintf("%s[%s]%s", BgGreen+White, pairType, Reset)
	}
	return fmt.Sprintf("%s[%s]%s", BgMagenta+White, pairType, Reset)
}

func formatDayThai(day string) string {
	switch day {
	case "sunday":
		return "วันอาทิตย์"
	case "monday":
		return "วันจันทร์"
	case "tuesday":
		return "วันอังคาร"
	case "wednesday1":
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
		return "ทุกวัน (ไม่ระบุ)"
	}
}

func truncateString(s string, maxLen int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen-3]) + "..."
}

func renderMethodologyGuide(reader *bufio.Reader) {
	fmt.Printf("\033[H\033[2J") // Clear screen
	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────────────────────┐%s\n", Cyan, Reset)
	fmt.Printf("%s│  📖  MIRACLE NAMES METHODOLOGY GUIDE - คู่มือกระบวนการคิดชื่อมงคล & หลักเกณฑ์จัดอันดับ  │%s\n", Bold+BgMagenta+White, Reset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────────────────────┤%s\n", Cyan, Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│  %s1. หลักการกาลกิณี (Kalakini Safety - 100%% Safe)%s                                   │\n", Bold+Yellow, Reset)
	fmt.Printf("│     • ระบบกรองกาลกิณีออก 100%% เสมอตามวันเกิด เพื่อความสบายใจสูงสุด                       │\n")
	fmt.Printf("│     • หากไม่ระบุวันเกิด ระบบจะอนุโลมให้ใช้งานได้กับทุกวันเกิด                              │\n")
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│  %s2. หลักการจัดอันดับ 3 ชั้น (3-Tier Hybrid Ranking Algorithm)%s                      │\n", Bold+Yellow, Reset)
	fmt.Printf("│     • %sTier 1 (ตรงเป้าหมาย 100%%):%s คัดเลือกชื่อที่มีอักษรทักษาเน้นที่ผู้ใช้เลือก (เช่น ศรี)   │\n", Green, Reset)
	fmt.Printf("│     • %sTier 2 (ทักษามงคลเสริม):%s หากชื่อตรงทักษาหลักหมด ระบบจะเติมชื่อทักษามงคลอื่นๆ   │\n", Yellow, Reset)
	fmt.Printf("│       (เช่น เดช, มนตรี, มูละ, อุตสาหะ, อายุ) ให้ครบ 100 รายชื่ออัตโนมัติ                   │\n")
	fmt.Printf("│     • %sTier 3 (พลังเลขศาสตร์มงคล):%s คัดเลือกชื่อเกรดมงคล (D10, D8, D5) ปลอดกาลกิณี     │\n", Cyan, Reset)
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("│  %s3. หลักการอนุโลม 2 พยางค์ขึ้นไป (Smart Flexibility Rule)%s                           │\n", Bold+Yellow, Reset)
	fmt.Printf("%s│     • ชื่อไทยนิยมใช้ 2 พยางค์ขึ้นไป หากบังคับผลรวมชื่อ+นามสกุลเป็น D10 อย่างเดียว       │%s\n", Reset, Reset)
	fmt.Printf("%s│       จะทำให้ตัวเลือกน้อยเกินไป ระบบจึงยืดหยุ่นให้ชื่อตัวเป็นเกรดมงคล (D) แล้วคำนวณ   │%s\n", Reset, Reset)
	fmt.Printf("│       ผลรวมกับนามสกุลให้เห็น เพื่อรักษาความงามและเป็นธรรมชาติของภาษาไทยไว้              │\n")
	fmt.Printf("│                                                                                        │\n")
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────────────────────┘%s\n", Cyan, Reset)

	fmt.Printf("\n%s[กด ENTER เพื่อกลับไปเมนูหลัก]: %s", Bold+Yellow, Reset)
	readInput(reader)
}
