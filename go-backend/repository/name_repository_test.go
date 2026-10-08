package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
	"go-backend/config"
	"go-backend/domain"
)

func TestCalculateTextSumsUsesDirectDictionaryForAe(t *testing.T) {
	dictionary := map[rune]characterNumerology{
		'แ': {satValue: 2, hasSat: true},
		'ก': {satValue: 1, shaValue: 15, hasSat: true, hasSha: true},
	}

	sat, sha, err := calculateTextSums("แก", dictionary)
	if err != nil {
		t.Fatalf("calculate sums: %v", err)
	}
	if sat != 3 || sha != 15 {
		t.Fatalf("expected direct dictionary แ=2/0 plus ก=1/15, got %d/%d", sat, sha)
	}
}

func TestCalculateTextSumsSupportsThaiCharacterMappedInOneDictionary(t *testing.T) {
	dictionary := map[rune]characterNumerology{
		'่': {satValue: 2, hasSat: true},
		'ฤ': {shaValue: 7, hasSha: true},
	}

	sat, sha, err := calculateTextSums("่ฤ", dictionary)
	if err != nil {
		t.Fatalf("expected sat-only and sha-only Thai characters to be supported: %v", err)
	}
	if sat != 2 || sha != 7 {
		t.Fatalf("expected missing dimensions to contribute zero, got %d/%d", sat, sha)
	}
}

func TestCalculateTextSumsRejectsCharacterAbsentFromBothDictionaries(t *testing.T) {
	_, _, err := calculateTextSums("🙂", map[rune]characterNumerology{})
	if !errors.Is(err, domain.ErrUnmappedCharacters) {
		t.Fatalf("expected unmapped-character error, got %v", err)
	}
}

func TestGovernanceSchemaUnavailableDetection(t *testing.T) {
	for _, code := range []pq.ErrorCode{"42P01", "3F000", "42703"} {
		if !isGovernanceSchemaUnavailable(&pq.Error{Code: code}) {
			t.Fatalf("expected PostgreSQL code %s to permit provisional fallback", code)
		}
	}
	if isGovernanceSchemaUnavailable(&pq.Error{Code: "42501"}) {
		t.Fatal("permission errors must not silently fall back to live calibration")
	}
}

func TestLowerHexSHA256Validation(t *testing.T) {
	if !isLowerHexSHA256(strings.Repeat("a", sha256.Size*2)) {
		t.Fatal("expected lowercase 64-character SHA-256 to be accepted")
	}
	for _, invalid := range []string{
		strings.Repeat("A", sha256.Size*2),
		strings.Repeat("g", sha256.Size*2),
		strings.Repeat("a", sha256.Size*2-1),
	} {
		if isLowerHexSHA256(invalid) {
			t.Fatalf("invalid SHA-256 was accepted: %q", invalid)
		}
	}
}

func TestSoundCorpusFingerprintRowUsesCanonicalJSONFieldOrder(t *testing.T) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(soundCorpusFingerprintRow{
		NameID:            7,
		THName:            "ชื่อ:\"ทดสอบ\"",
		PhoneticVersion:   domain.MissingPhoneticVersion,
		EuphonyScore:      80,
		PronunciationEase: 70,
		RhythmScore:       60,
	})
	if err != nil {
		t.Fatalf("marshal fingerprint row: %v", err)
	}
	const want = `{"name_id":7,"thname":"ชื่อ:\"ทดสอบ\"","phonetic_version":"<missing>","euphony_score":80,"pronunciation_ease":70,"rhythm_score":60}` + "\n"
	if encoded.String() != want {
		t.Fatalf("fingerprint JSON contract changed:\n got %s\nwant %s", encoded.String(), want)
	}
}

func TestSemanticEmbeddingInputMatchesVersionedTemplate(t *testing.T) {
	if got := semanticEmbeddingInput("  ผู้นำที่สง่างาม  "); got != "ความหมาย: ผู้นำที่สง่างาม" {
		t.Fatalf("semantic input template %q produced %q", domain.SemanticInputTemplateVersion, got)
	}
}

func TestApprovedCalibrationMustMatchCurrentCorpus(t *testing.T) {
	fingerprint := strings.Repeat("a", sha256.Size*2)
	profile := domain.SoundCalibrationProfile{
		Source:                  "approved_snapshot",
		Approved:                true,
		EvaluationVersion:       "thai_name_human_eval_v1",
		CorpusFingerprintSHA256: fingerprint,
		Versions: map[string]domain.SoundCalibrationStats{
			"v1": {Count: 100, Mean: 80, StdDev: 5},
		},
	}
	current := soundCorpusIdentity{
		FingerprintSHA256: fingerprint,
		VersionCounts:     map[string]int64{"v1": 100, "v2": 20},
		TotalCount:        120,
	}
	if approvedProfileMatchesCorpus(profile, current) {
		t.Fatal("partial approved snapshot must not be used")
	}

	profile.Versions["v2"] = domain.SoundCalibrationStats{Count: 20, Mean: 70, StdDev: 4}
	if !approvedProfileMatchesCorpus(profile, current) {
		t.Fatal("complete approved snapshot should cover the current corpus versions")
	}

	profile.Versions["v2"] = domain.SoundCalibrationStats{Count: 20, Mean: 70, StdDev: 0}
	if approvedProfileMatchesCorpus(profile, current) {
		t.Fatal("snapshot entry with invalid standard deviation must fail coverage")
	}
	profile.Versions["v2"] = domain.SoundCalibrationStats{Count: 20, Mean: 70, StdDev: 4}
	profile.EvaluationVersion = ""
	if approvedProfileMatchesCorpus(profile, current) {
		t.Fatal("snapshot without a human evaluation version must not be marked approved")
	}
	profile.EvaluationVersion = "thai_name_human_eval_v1"
	current.FingerprintSHA256 = strings.Repeat("b", sha256.Size*2)
	if approvedProfileMatchesCorpus(profile, current) {
		t.Fatal("snapshot fingerprint must match the current eligible corpus")
	}
}

func TestCloneSoundCalibrationProfileDefensivelyCopiesVersions(t *testing.T) {
	profile := domain.SoundCalibrationProfile{
		Source:   "approved_snapshot",
		Approved: true,
		Versions: map[string]domain.SoundCalibrationStats{
			"v1": {Count: 10, Mean: 80, StdDev: 5},
		},
	}

	cloned := cloneSoundCalibrationProfile(profile)
	cloned.Versions["v1"] = domain.SoundCalibrationStats{Count: 1}
	if profile.Versions["v1"].Count != 10 {
		t.Fatal("profile clone shares a mutable versions map")
	}
}

func TestCandidateEligibilityClauseAppliesDayCanonicalAeAndChildTitleGuards(t *testing.T) {
	clause := candidateEligibilityClause("monday")
	for _, required := range []string{
		"k_monday = false",
		"STRPOS(n.thname, 'แ')",
		"n.sat_sum =",
		"LEFT JOIN sat_nums",
		"n.sha_sum =",
		"LEFT JOIN sha_nums",
		"n.thname NOT LIKE 'เด็กหญิง%'",
		"n.thname NOT LIKE 'เด็กชาย%'",
		"AND translate(",
		"SELECT char_key FROM sat_nums",
		"UNION",
		"SELECT char_key FROM sha_nums",
		") = ''",
	} {
		if !strings.Contains(clause, required) {
			t.Fatalf("candidate safety clause is missing %q: %s", required, clause)
		}
	}

	withoutDay := candidateEligibilityClause("")
	if strings.Contains(withoutDay, "k_monday") || !strings.Contains(withoutDay, "STRPOS(n.thname, 'แ')") {
		t.Fatalf("canonical anomaly guard must apply even without a day filter: %s", withoutDay)
	}
	for _, broadPrefix := range []string{"นาง%", "นาย%", "พระ%"} {
		if strings.Contains(withoutDay, "NOT LIKE '"+broadPrefix+"'") {
			t.Fatalf("ambiguous prefix %q must remain review-only, not auto-filtered: %s", broadPrefix, withoutDay)
		}
	}
}

func TestNameRepositoryGetSumsByTHNameUsesCanonicalDictionary(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	defer db.Close()

	const name = "แก้วกรต"
	var directSat, directSha int
	err = db.QueryRowContext(context.Background(), `
		SELECT
			COALESCE((
				SELECT SUM(COALESCE(s.sat_value, 0))
				FROM regexp_split_to_table($1, '') AS chars(char_key)
				LEFT JOIN sat_nums AS s ON s.char_key = chars.char_key
			), 0),
			COALESCE((
				SELECT SUM(COALESCE(sh.sha_value, 0))
				FROM regexp_split_to_table($1, '') AS chars(char_key)
				LEFT JOIN sha_nums AS sh ON sh.char_key = chars.char_key
			), 0)
	`, name).Scan(&directSat, &directSha)
	if err != nil {
		t.Fatalf("calculate independent dictionary oracle: %v", err)
	}

	repo := NewNameRepository(db)
	sat, sha, _, err := repo.GetSumsByTHName(context.Background(), name)
	if err != nil {
		t.Fatalf("get canonical known-name sums: %v", err)
	}
	if sat != directSat || sha != directSha {
		t.Fatalf("known-name lookup returned stale sums %d/%d; direct dictionary is %d/%d", sat, sha, directSat, directSha)
	}
}
