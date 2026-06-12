package collector

import (
	"context"
	"syscall"

	"moon/internal/entity"
	"moon/pkg/types"
)

func NewDiskCollector() entity.Collector {
	return func(ctx context.Context, m *entity.Metrics) error {
		var stat syscall.Statfs_t
		if err := syscall.Statfs("/", &stat); err != nil {
			return err
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free

		space := types.GiB(total / (1024 * 1024 * 1024))
		usage := types.Percent(float64(used) / float64(total) * 100)

		m.Set("disk", &entity.Disk{
			Space: space,
			Usage: usage,
		})

		return nil
	}
}
