package app

import (
	"moon/internal/core"
	"moon/internal/entity"
)

func Run(machine *entity.Machine, collectors []core.Collector) {
	errors := make(chan error)

	for _, collector := range collectors {
		go func() {
			for {
				if err := collector(machine); err != nil {
					errors <- err
				}
			}
		}()
	}
}
