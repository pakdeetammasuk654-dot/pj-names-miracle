package domain

import "context"

// PipelineContext wraps context and data passed through pipeline filters
type PipelineContext[T any] struct {
	Ctx  context.Context
	Data T
}

// Filter defines a single processing step (Filter) in the pipeline
type Filter[T any] interface {
	Execute(pCtx *PipelineContext[T]) error
}

// Pipeline represents a sequential collection of Filters (Pipe)
type Pipeline[T any] struct {
	filters []Filter[T]
}

// NewPipeline creates a new pipeline with given filters
func NewPipeline[T any](filters ...Filter[T]) *Pipeline[T] {
	return &Pipeline[T]{
		filters: filters,
	}
}

// AddFilter registers a filter to the pipeline
func (p *Pipeline[T]) AddFilter(f Filter[T]) *Pipeline[T] {
	p.filters = append(p.filters, f)
	return p
}

// Execute runs all filters sequentially. If any filter returns an error, execution stops.
func (p *Pipeline[T]) Execute(pCtx *PipelineContext[T]) error {
	for _, filter := range p.filters {
		if err := filter.Execute(pCtx); err != nil {
			return err
		}
	}
	return nil
}
