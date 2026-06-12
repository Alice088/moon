package analyzer

import (
	"context"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewRAMPeak(thresholdPct types.Percent) entity.Analyzer {
	return func(ctx context.Context, m *entity.Metrics) error {
		if thresholdPct <= 0 {
			return nil
		}
		ram, ok := m.Get("ram").(*entity.RAM)
		if !ok {
			return nil
		}
		if ram.Usage > thresholdPct {
			m.Set("alert", entity.Alert{
				Type:    "ram_peak",
				Message: "ram usage exceeds threshold",
				Data: map[string]any{
					"usage":     ram.Usage,
					"threshold": thresholdPct,
				},
			})
		}
		return nil
	}
}
