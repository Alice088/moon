package entity

import (
	"context"
	"time"
)

type Collector func(ctx context.Context, metrics *Metrics) error

type Pipeline struct {
	Input      *Metrics      `json:"input"`
	Output     chan *Metrics `json:"output"`
	Errors     chan error    `json:"errors"`
	Collectors []Collector   `json:"collectors"`
	Interval   time.Duration `json:"interval"`
}

func NewPipeline(collectors []Collector) *Pipeline {
	return &Pipeline{
		Input:      NewMetrics("pipeline"),
		Output:     make(chan *Metrics, 100),
		Errors:     make(chan error, 100),
		Collectors: collectors,
		Interval:   time.Second,
	}
}

func (p *Pipeline) Run(ctx context.Context) (<-chan *Metrics, <-chan error) {
	go func() {
		defer close(p.Output)
		defer close(p.Errors)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, c := range p.Collectors {
					if err := c(ctx, p.Input); err != nil {
						select {
						case p.Errors <- err:
						case <-ctx.Done():
							return
						}
					}
				}

				select {
				case p.Output <- p.Input:
				case <-ctx.Done():
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(p.Interval):
				}
			}
		}
	}()

	return p.Output, p.Errors
}
