// Command name-import prepares a reviewed batch of Thai names for enrichment.
// It never writes directly to public.names_miracle.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"go-backend/config"
)

const maxExamples = 25

type sourceName struct {
	line int
	name string
}

type rejectedName struct {
	Line   int    `json:"line"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type preflightReport struct {
	InputFile        string         `json:"input_file"`
	SourceSHA256     string         `json:"source_sha256"`
	InputLines       int            `json:"input_lines"`
	Accepted         int            `json:"accepted"`
	AutoNormalized   int            `json:"auto_normalized"`
	DuplicateInFile  int            `json:"duplicate_in_file"`
	Invalid          int            `json:"invalid"`
	AcceptedExamples []string       `json:"accepted_examples"`
	RejectedExamples []rejectedName `json:"rejected_examples"`
	GeneratedAt      string         `json:"generated_at"`
}

func main() {
	mode := flag.String("mode", "preflight", "preflight or stage")
	input := flag.String("input", "", "UTF-8 .txt (one name per line) or .csv file")
	batchLabel := flag.String("batch", "", "unique batch label, required for stage")
	createdBy := flag.String("created-by", "", "operator name, required for stage")
	apply := flag.Bool("apply", false, "required with -mode stage; otherwise no database write is made")
	flag.Parse()

	if strings.TrimSpace(*input) == "" {
		fatal(errors.New("-input is required"))
	}
	names, sourceHash, lines, err := readNames(*input)
	if err != nil {
		fatal(err)
	}
	report, accepted := buildReport(*input, sourceHash, lines, names)
	printJSON(report)
	if report.Accepted == 0 {
		fatal(errors.New("no valid Thai names remain after preflight"))
	}

	switch *mode {
	case "preflight":
		return
	case "stage":
		if !*apply {
			fatal(errors.New("refusing to write: rerun with -apply after reviewing the preflight report"))
		}
		if strings.TrimSpace(*batchLabel) == "" || strings.TrimSpace(*createdBy) == "" {
			fatal(errors.New("-batch and -created-by are required for stage"))
		}
		if err := stageBatch(strings.TrimSpace(*batchLabel), strings.TrimSpace(*createdBy), sourceHash, accepted); err != nil {
			fatal(err)
		}
		fmt.Printf("Staged %d names in batch %q. No rows were added to names_miracle.\n", len(accepted), *batchLabel)
	default:
		fatal(fmt.Errorf("unsupported -mode %q", *mode))
	}
}

func readNames(path string) ([]sourceName, string, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", 0, fmt.Errorf("read input: %w", err)
	}
	digest := sha256.Sum256(raw)
	hash := hex.EncodeToString(digest[:])
	if strings.EqualFold(filepath.Ext(path), ".csv") {
		return readCSV(raw, hash)
	}

	var names []sourceName
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	// Thai compound names and surrounding CSV-like text can exceed Scanner's default.
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		names = append(names, sourceName{line: line, name: strings.TrimSpace(scanner.Text())})
	}
	if err := scanner.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("read input lines: %w", err)
	}
	return names, hash, line, nil
}

func readCSV(raw []byte, hash string) ([]sourceName, string, int, error) {
	r := csv.NewReader(strings.NewReader(string(raw)))
	r.FieldsPerRecord = -1
	var names []sourceName
	line := 0
	for {
		record, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, "", 0, fmt.Errorf("read CSV: %w", err)
		}
		line++
		if len(record) > 0 {
			names = append(names, sourceName{line: line, name: strings.TrimSpace(record[0])})
		}
	}
	return names, hash, line, nil
}

func buildReport(input, hash string, lines int, names []sourceName) (preflightReport, []sourceName) {
	report := preflightReport{InputFile: input, SourceSHA256: hash, InputLines: lines, GeneratedAt: time.Now().UTC().Format(time.RFC3339)}
	seen := make(map[string]struct{})
	accepted := make([]sourceName, 0, len(names))
	for _, item := range names {
		normalized := normalizeName(item.name)
		if normalized != item.name {
			report.AutoNormalized++
			item.name = normalized
		}
		if reason := validateName(item.name); reason != "" {
			report.Invalid++
			appendRejected(&report, item, reason)
			continue
		}
		if _, duplicate := seen[item.name]; duplicate {
			report.DuplicateInFile++
			appendRejected(&report, item, "duplicate_in_input")
			continue
		}
		seen[item.name] = struct{}{}
		accepted = append(accepted, item)
		if len(report.AcceptedExamples) < maxExamples {
			report.AcceptedExamples = append(report.AcceptedExamples, item.name)
		}
	}
	report.Accepted = len(accepted)
	sort.Strings(report.AcceptedExamples)
	return report, accepted
}

// normalizeName removes Unicode format marks accidentally introduced by copy
// and paste. These marks have no pronunciation or numerology value and must
// not create a second spelling of the same Thai name. The source file itself
// is never modified.
func normalizeName(name string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF':
			return -1
		default:
			return r
		}
	}, strings.TrimSpace(name))
}

func validateName(name string) string {
	if name == "" {
		return "blank"
	}
	if utf8RuneCount(name) < 2 {
		return "too_short"
	}
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsDigit(r) || unicode.IsLetter(r) && (r < 0x0E00 || r > 0x0E7F) {
			return "contains_non_thai_or_whitespace"
		}
		if r < 0x0E00 || r > 0x0E7F {
			return "contains_unsupported_character"
		}
	}
	return ""
}

func utf8RuneCount(s string) int {
	return len([]rune(s))
}

func appendRejected(report *preflightReport, item sourceName, reason string) {
	if len(report.RejectedExamples) < maxExamples {
		report.RejectedExamples = append(report.RejectedExamples, rejectedName{Line: item.line, Name: item.name, Reason: reason})
	}
}

func stageBatch(label, createdBy, sourceHash string, names []sourceName) error {
	db, err := config.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var batchID int64
	err = tx.QueryRow(`
		INSERT INTO data_governance.name_import_batches (batch_label, source_sha256, source_row_count, created_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (batch_label) DO UPDATE
		SET source_sha256 = EXCLUDED.source_sha256,
			source_row_count = EXCLUDED.source_row_count
		RETURNING batch_id`, label, sourceHash, len(names), createdBy).Scan(&batchID)
	if err != nil {
		return fmt.Errorf("create staging batch (apply migration 000024 first): %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO data_governance.name_import_staging (batch_id, source_line, thname, normalized_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (batch_id, normalized_name) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, item := range names {
		if _, err := stmt.Exec(batchID, item.line, item.name, item.name); err != nil {
			return fmt.Errorf("stage %q: %w", item.name, err)
		}
	}
	return tx.Commit()
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(b))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "name-import:", err)
	os.Exit(1)
}
