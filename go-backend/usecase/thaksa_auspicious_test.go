package usecase

import (
	"strings"
	"testing"
	"time"
)

func TestCalculateTaksaIncludesRoleMeaningAndAuspiciousGuidance(t *testing.T) {
	current := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	result := CalculateTaksa("1982-07-10", "02:59", current)

	if len(result.AllRoles) != 8 {
		t.Fatalf("all roles = %d, want 8", len(result.AllRoles))
	}
	for _, role := range result.AllRoles {
		if strings.TrimSpace(role.Meaning) == "" {
			t.Fatalf("role %q has empty meaning", role.RoleName)
		}
		// The full day label includes its planet number; guidance intentionally uses the readable day name.
		dayName := strings.Split(role.DayName, " (")[0]
		if !strings.Contains(role.AuspiciousGuidance, dayName) {
			t.Fatalf("guidance %q does not identify day %q", role.AuspiciousGuidance, role.DayName)
		}
	}
}
