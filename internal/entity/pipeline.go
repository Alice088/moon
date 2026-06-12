package entity

import (
	"context"
	"sync"
)

type Collector func(ctx context.Context, metrics *Metrics) error

type Pipeline struct {
	Input      *Metrics      `json:"input"`
	Output     chan *Metrics `json:"output"`
	Errors     chan error    `json:"errors"`
	Collectors []Collector   `json:"collectors"`
}

func NewPipeline(collectors []Collector) *Pipeline {
	return &Pipeline{
		Input:      NewMetrics("pipeline"),
		Output:     make(chan *Metrics, 100),
		Errors:     make(chan error, 100),
		Collectors: collectors,
	}
}

func (p *Pipeline) Run(ctx context.Context) (<-chan *Metrics, <-chan error) {
	var wg sync.WaitGroup

	for _, c := range p.Collectors {
		wg.Add(1)
		go func(collector Collector) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := collector(ctx, p.Input); err != nil {
						select {
						case p.Errors <- err:
						case <-ctx.Done():
							return
						}
					}

					select {
					case p.Output <- p.Input:
					case <-ctx.Done():
						return
					}
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(p.Output)
		close(p.Errors)
	}()

	return p.Output, p.Errors
}
