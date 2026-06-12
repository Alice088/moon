package analyzer

import (
	"context"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewCPUPeak(thresholdPct types.Percent) entity.Analyzer {
	return func(ctx context.Context, m *entity.Metrics) error {
		if thresholdPct <= 0 {
			return nil
		}
		cpu, ok := m.Get("cpu").(*entity.CPU)
		if !ok {
			return nil
		}
		usagePct := types.Percent(100 - cpu.Average.Idle)
		if usagePct > thresholdPct {
			m.Set("alert", entity.Alert{
				Type:    "cpu_peak",
				Message: "cpu usage exceeds threshold",
				Data: map[string]any{
					"usage":     usagePct,
					"threshold": thresholdPct,
				},
			})
		}
		return nil
	}
}
