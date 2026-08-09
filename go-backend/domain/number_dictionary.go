package domain

import "strings"

type NumberMeaning struct {
	PairNumber    string `json:"pairnumber"`
	PairType      string `json:"pairtype"`      // D10, D8, D5, R10, R7, R5
	PairPoint     int    `json:"pairpoint"`     // Points e.g. 80, -50
	MiracleDesc   string `json:"miracledesc"`   // Short summary
	MiracleDetail string `json:"miracledetail"` // Detail
}

// IsGood follows the domain grading: every D-type pair is auspicious. A D5 row
// may still carry a low or negative point value, which lowers ranking score
// without excluding the name from results.
func (n NumberMeaning) IsGood() bool {
	pairType := strings.ToUpper(strings.TrimSpace(n.PairType))
	return strings.HasPrefix(pairType, "D")
}
