package analyzer

import (
	"context"
	"fmt"

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
			return fmt.Errorf(
				"cpu peak: usage %.1f%% exceeds threshold %.1f%%",
				usagePct, thresholdPct,
			)
		}
		return nil
	}
}
