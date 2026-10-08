package domain_test

import (
	"testing"

	"go-backend/domain"
)

func TestCombineMeanings(t *testing.T) {
	tests := []struct {
		name     string
		m1       string
		m2       string
		expected string
	}{
		{
			name:     "both empty",
			m1:       "",
			m2:       "",
			expected: "",
		},
		{
			name:     "m1 empty",
			m1:       "",
			m2:       "ผู้มั่นคงดั่งกำแพง",
			expected: "ผู้มั่นคงดั่งกำแพง",
		},
		{
			name:     "m2 empty",
			m1:       "หญิงสาวผู้มีความกล้าหาญ",
			m2:       "",
			expected: "หญิงสาวผู้มีความกล้าหาญ",
		},
		{
			name:     "combines name and surname meanings",
			m1:       "หญิงสาวผู้มีความกล้าหาญ",
			m2:       "ผู้มั่นคงดั่งกำแพง",
			expected: "หญิงสาวผู้มีความกล้าหาญ มั่นคงดั่งกำแพง",
		},
		{
			name:     "m2 already starts with และ",
			m1:       "ผู้มีความสุข",
			m2:       "และมีความเจริญ",
			expected: "ผู้มีความสุข มีความเจริญ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CombineMeanings(tt.m1, tt.m2)
			if got != tt.expected {
				t.Errorf("CombineMeanings(%q, %q) = %q; want %q", tt.m1, tt.m2, got, tt.expected)
			}
		})
	}
}
