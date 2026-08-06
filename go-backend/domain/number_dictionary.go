package domain

import "strings"

type NumberMeaning struct {
	PairNumber    string `json:"pairnumber"`
	PairType      string `json:"pairtype"`      // D10, D8, D5, R10, R7, R5
	PairPoint     int    `json:"pairpoint"`     // Points e.g. 80, -50
	MiracleDesc   string `json:"miracledesc"`   // Short summary
	MiracleDetail string `json:"miracledetail"` // Detail
}

// IsGood returns true if pairtype starts with 'D' (D10, D8, D5)
func (n NumberMeaning) IsGood() bool {
	pairType := strings.ToUpper(strings.TrimSpace(n.PairType))
	return strings.HasPrefix(pairType, "D")
}
