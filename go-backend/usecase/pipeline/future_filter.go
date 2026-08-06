package pipeline

import (
	"log"

	"go-backend/domain"
)

// FutureExtensionFilter is a ready-to-use placeholder for future logic (e.g. Day of Birth / Kalakini / Meanings)
type FutureExtensionFilter struct{}

func NewFutureExtensionFilter() *FutureExtensionFilter {
	return &FutureExtensionFilter{}
}

func (f *FutureExtensionFilter) Execute(pCtx *domain.PipelineContext[*domain.NameSearchPayload]) error {
	// Ready to plug in next factors when defined!
	log.Printf("[Pipeline] Step: Future Extension Filter passed (Ready for next rules)")
	return nil
}
