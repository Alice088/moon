package analyzer

import (
	"context"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewDiskPeak(thresholdPct types.Percent) entity.Analyzer {
	return func(ctx context.Context, m *entity.Metrics) error {
		if thresholdPct <= 0 {
			return nil
		}
		disk, ok := m.Get("disk").(*entity.Disk)
		if !ok {
			return nil
		}
		if disk.Usage > thresholdPct {
			m.Set("alert", entity.Alert{
				Type:    "disk_peak",
				Message: "disk space usage exceeds threshold",
				Data: map[string]any{
					"usage":     disk.Usage,
					"threshold": thresholdPct,
				},
			})
		}
		return nil
	}
}
