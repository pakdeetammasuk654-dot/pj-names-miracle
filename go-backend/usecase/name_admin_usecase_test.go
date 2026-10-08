package usecase

import (
	"context"
	"testing"

	"go-backend/domain"
)

type mockNameAdminRepo struct {
	names map[string]int64
}

func (m *mockNameAdminRepo) GetDictionary(ctx context.Context) (map[rune]int, map[rune]int, error) {
	sat := map[rune]int{
		'เ': 2, 'จ': 6, 'น': 5, '่': 1, 'ณ': 5, 'ั': 4, 'ฐ': 9, 'ก': 1, 'ร': 4,
	}
	sha := map[rune]int{
		'เ': 6, 'จ': 8, 'น': 10, '่': 0, 'ณ': 8, 'ั': 4, 'ฐ': 8, 'ก': 2, 'ร': 6,
	}
	return sat, sha, nil
}

func (m *mockNameAdminRepo) GetKalakiniRules(ctx context.Context) (map[string]map[rune]bool, error) {
	rules := map[string]map[rune]bool{
		"k_sunday":       {'ศ': true, 'ษ': true, 'ส': true, 'ห': true, 'ฬ': true, 'ฮ': true},
		"k_monday":       {'เ': true, 'แ': true, 'โ': true, 'ใ': true, 'ไ': true},
		"k_tuesday":      {'ก': true, 'ข': true, 'ค': true, 'ฆ': true, 'ง': true},
		"k_wednesday1":   {'จ': true, 'ฉ': true, 'ช': true, 'ซ': true, 'ฌ': true, 'ญ': true},
		"k_wednesday2":   {'บ': true, 'ป': true, 'ผ': true, 'ฝ': true, 'พ': true, 'ฟ': true, 'ภ': true, 'ม': true},
		"k_thursday":     {'ด': true, 'ต': true, 'ถ': true, 'ท': true, 'ธ': true, 'น': true},
		"k_friday":       {'ย': true, 'ร': true, 'ล': true, 'ว': true},
		"k_saturday":     {'ฎ': true, 'ฏ': true, 'ฐ': true, 'ฑ': true, 'ฒ': true, 'ณ': true},
	}
	return rules, nil
}

func (m *mockNameAdminRepo) ExistsTHName(ctx context.Context, thname string) (bool, int64, error) {
	id, ok := m.names[thname]
	return ok, id, nil
}

func (m *mockNameAdminRepo) InsertCompleteName(ctx context.Context, item *domain.AdminNameItem, vectorJSON string) (int64, error) {
	id := int64(len(m.names) + 1001)
	m.names[item.THName] = id
	return id, nil
}

func (m *mockNameAdminRepo) GetRecentNames(ctx context.Context, search string, limit int, offset int) ([]domain.AdminNameItem, int, error) {
	return []domain.AdminNameItem{}, len(m.names), nil
}

func (m *mockNameAdminRepo) GetTotalNamesCount(ctx context.Context) (int, error) {
	return len(m.names), nil
}

func TestNameAdminUseCase_AnalyzeName(t *testing.T) {
	mockRepo := &mockNameAdminRepo{
		names: map[string]int64{
			"มีอยู่แล้ว": 999,
		},
	}
	uc := NewNameAdminUseCase(mockRepo)

	ctx := context.Background()

	// Test 1: Analyze new name "เจเน่"
	res, err := uc.AnalyzeName(ctx, "เจเน่", false, "ค")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.THName != "เจเน่" {
		t.Errorf("expected THName 'เจเน่', got %s", res.THName)
	}
	if res.SatSum != 16 {
		t.Errorf("expected SatSum 16, got %d", res.SatSum)
	}
	if res.ShaSum != 30 {
		t.Errorf("expected ShaSum 30, got %d", res.ShaSum)
	}
	if res.AlreadyExists {
		t.Errorf("expected AlreadyExists false, got true")
	}

	// Test 2: Analyze existing name
	resDup, err := uc.AnalyzeName(ctx, "มีอยู่แล้ว", false, "ค")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resDup.AlreadyExists {
		t.Errorf("expected AlreadyExists true, got false")
	}
	if resDup.ExistingID != 999 {
		t.Errorf("expected ExistingID 999, got %d", resDup.ExistingID)
	}
}

func TestFormatSumPairs(t *testing.T) {
	tests := []struct {
		total int
		want  []string
	}{
		{0, []string{}},
		{9, []string{"09"}},
		{16, []string{"16"}},
		{111, []string{"11", "11"}},
		{1234, []string{"12", "23", "34"}},
	}

	for _, tt := range tests {
		got := formatSumPairs(tt.total)
		if len(got) != len(tt.want) {
			t.Errorf("formatSumPairs(%d) count = %d, want %d", tt.total, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("formatSumPairs(%d)[%d] = %s, want %s", tt.total, i, got[i], tt.want[i])
			}
		}
	}
}
