package hook

import (
	"context"
	"sync"

	"moon/internal/entity"
)

type Runner struct {
	hooks   []Hook
	workers int
}

func NewRunner(hooks []Hook, workers int) *Runner {
	return &Runner{hooks: hooks, workers: workers}
}

func (r *Runner) Run(ctx context.Context, input <-chan *entity.Metrics) <-chan error {
	errs := make(chan error, 100)
	var wg sync.WaitGroup

	for i := 0; i < r.workers; i++ {
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
					for _, h := range r.hooks {
						if err := h(ctx, m); err != nil {
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
