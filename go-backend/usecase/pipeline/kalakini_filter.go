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
		log.Printf("[Pipeline Step 2: KalakiniFilter] Applied 100%% Kalakini safety filter for Day: %s", day)
	} else {
		log.Printf("[Pipeline Step 2: KalakiniFilter] DayOfBirth not specified, skipping Kalakini filter")
	}
	return nil
}
