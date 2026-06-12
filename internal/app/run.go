package app

import (
	"moon/internal/core"
	"moon/internal/entity"
)

func Run(machine *entity.Machine, monitor []core.CollectFunc) {
	errors := make(chan error)

	for _, fn := range monitor {
		go func() {
			for {
				if err := fn(machine); err != nil {
					errors <- err
				}
			}
		}()
	}
}
