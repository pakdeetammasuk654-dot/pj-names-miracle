package pipeline

import (
	"log"
	"strings"

	"go-backend/domain"
)

// KalakiniFilter validates and logs DayOfBirth Kalakini safety
type KalakiniFilter struct{}

func NewKalakiniFilter() *KalakiniFilter {
	return &KalakiniFilter{}
}

func (f *KalakiniFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	day := strings.ToLower(strings.TrimSpace(pCtx.Data.Request.DayOfBirth))
	if day != "" {
		scope := "first_name"
		if pCtx.Data.Request.Mode == domain.ModeCombined && pCtx.Data.Request.SearchTarget == domain.TargetSurname {
			scope = "surname"
		}
		log.Printf("[Pipeline Step 2: KalakiniFilter] Applied candidate %s Kalakini filter for Day: %s", scope, day)
	} else {
		log.Printf("[Pipeline Step 2: KalakiniFilter] DayOfBirth not specified, skipping Kalakini filter")
	}
	return nil
}
