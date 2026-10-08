package main

import (
	"reflect"
	"testing"
)

func TestFormatSumPairs(t *testing.T) {
	tests := []struct {
		name  string
		total int
		want  []string
	}{
		{name: "non-positive", total: 0, want: []string{}},
		{name: "single digit", total: 6, want: []string{"06"}},
		{name: "two digits", total: 41, want: []string{"41"}},
		{name: "three digits", total: 110, want: []string{"11", "10"}},
		{name: "four digits", total: 1234, want: []string{"12", "23", "34"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatSumPairs(tt.total); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("formatSumPairs(%d) = %v, want %v", tt.total, got, tt.want)
			}
		})
	}
}
