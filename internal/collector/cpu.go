package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewCPUCollector() entity.Collector {
	var prev [8]uint64
	var first bool

	return func(ctx context.Context, m *entity.Metrics) error {
		f, err := os.Open("/proc/stat")
		if err != nil {
			return fmt.Errorf("open /proc/stat: %w", err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Scan()
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 8 || fields[0] != "cpu" {
			return fmt.Errorf("unexpected /proc/stat format")
		}

		var cur [8]uint64
		for i := 0; i < 8; i++ {
			cur[i], _ = strconv.ParseUint(fields[1+i], 10, 64)
		}

		if !first {
			prev = cur
			first = true
			return nil
		}

		prevIdle := prev[3] + prev[4]
		curIdle := cur[3] + cur[4]
		prevTotal := prev[0] + prev[1] + prev[2] + prev[3] + prev[4] + prev[5] + prev[6] + prev[7]
		curTotal := cur[0] + cur[1] + cur[2] + cur[3] + cur[4] + cur[5] + cur[6] + cur[7]

		totalDelta := curTotal - prevTotal
		if totalDelta == 0 {
			prev = cur
			return nil
		}

		idleDelta := curIdle - prevIdle

		m.Set("cpu", &entity.CPU{
			Average: entity.Usage{
				User:   types.Percent(float64(cur[0]-prev[0]) / float64(totalDelta) * 100),
				Nice:   types.Percent(float64(cur[1]-prev[1]) / float64(totalDelta) * 100),
				System: types.Percent(float64(cur[2]-prev[2]) / float64(totalDelta) * 100),
				Idle:   types.Percent(float64(idleDelta) / float64(totalDelta) * 100),
				IOWait: types.Percent(float64(cur[4]-prev[4]) / float64(totalDelta) * 100),
				Steal:  types.Percent(float64(cur[7]-prev[7]) / float64(totalDelta) * 100),
			},
			Cores:     1,
			Timestamp: time.Now(),
		})

		prev = cur
		return nil
	}
}
