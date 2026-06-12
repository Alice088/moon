package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewRAMCollector() entity.Collector {
	return func(ctx context.Context, m *entity.Metrics) error {
		f, err := os.Open("/proc/meminfo")
		if err != nil {
			return fmt.Errorf("open /proc/meminfo: %w", err)
		}
		defer f.Close()

		var memTotal, memAvailable uint64
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "MemTotal:"):
				memTotal, _ = strconv.ParseUint(strings.Fields(line)[1], 10, 64)
			case strings.HasPrefix(line, "MemAvailable:"):
				memAvailable, _ = strconv.ParseUint(strings.Fields(line)[1], 10, 64)
			}
			if memTotal > 0 && memAvailable > 0 {
				break
			}
		}

		if memTotal == 0 {
			return fmt.Errorf("memTotal not found in /proc/meminfo")
		}

		usage := types.Percent(float64(memTotal-memAvailable) / float64(memTotal) * 100)

		m.Set("ram", &entity.RAM{
			Usage: usage,
		})

		return nil
	}
}
