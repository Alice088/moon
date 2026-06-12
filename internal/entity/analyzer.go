package entity

import (
	"context"
	"sync"
)

type Analyzer func(ctx context.Context, metrics *Metrics) error

type AnalyzerPool struct {
	analyzers []Analyzer
	workers   int
}

func NewAnalyzerPool(analyzers []Analyzer, workers int) *AnalyzerPool {
	return &AnalyzerPool{
		analyzers: analyzers,
		workers:   workers,
	}
}

func (p *AnalyzerPool) Run(ctx context.Context, input <-chan *Metrics) <-chan error {
	errs := make(chan error, 100)
	var wg sync.WaitGroup

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-input:
					if !ok {
						return
					}
					snapshot := m.GetAll()
					for _, a := range p.analyzers {
						copy := NewMetrics("analyzer")
						for k, v := range snapshot {
							copy.Set(k, v)
						}
						if err := a(ctx, copy); err != nil {
							select {
							case errs <- err:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	return errs
}
