package dispatcher

import (
	"context"

	"moon/internal/entity"
)

func Run(ctx context.Context, input <-chan *entity.Metrics, notifiers []entity.Notifier) <-chan error {
	errs := make(chan error, 100)
	go func() {
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-input:
				if !ok {
					return
				}
				alert, ok := m.Get("alert").(entity.Alert)
				if !ok {
					continue
				}
				for _, n := range notifiers {
					if err := n.Send(ctx, alert); err != nil {
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
	return errs
}
