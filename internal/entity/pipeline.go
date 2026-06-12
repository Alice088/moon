package entity

import (
	"context"
	"sync"
)

type Collector func(ctx context.Context, metrics *Metrics) error

type Pipeline struct {
	input      *Metrics
	output     chan *Metrics
	errors     chan error
	collectors []Collector
}

func NewPipeline(collectors []Collector) *Pipeline {
	return &Pipeline{
		input:      NewMetrics("pipeline"),
		output:     make(chan *Metrics, 100),
		errors:     make(chan error, 100),
		collectors: collectors,
	}
}

func (p *Pipeline) Run(ctx context.Context) (<-chan *Metrics, <-chan error) {
    var wg sync.WaitGroup
    
    for _, c := range p.collectors {
        wg.Add(1)
        go func(collector Collector) {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    if err := collector(ctx, p.input); err != nil {
                        select {
                        case p.errors <- err:
                        case <-ctx.Done():
                            return
                        }
                    }
                    
                    select {
                    case p.output <- p.input:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }(c)
    }
    
    go func() {
        wg.Wait()
        close(p.output)
        close(p.errors)
    }()
    
    return p.output, p.errors
}