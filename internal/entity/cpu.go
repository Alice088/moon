package entity

import (
	"context"
	"fmt"

	"moon/pkg/types"
	"time"
)

type Usage struct {
	User   types.Percent `json:"user"`
	System types.Percent `json:"system"`
	Idle   types.Percent `json:"idle"`
	IOWait types.Percent `json:"iowait"`
	Steal  types.Percent `json:"steal"`
	Nice   types.Percent `json:"nice"`
}

type CPU struct {
	Usage     []Usage   `json:"usage"`
	Average   Usage     `json:"average"`
	Cores     int       `json:"cores"`
	Model     string    `json:"model"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Timestamp time.Time `json:"timestamp"`
}

func NewCPUPeakAnalyzer(thresholdPct types.Percent) Analyzer {
	return func(ctx context.Context, m *Metrics) error {
		if thresholdPct <= 0 {
			return nil
		}
		cpu, ok := m.Get("cpu").(*CPU)
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
