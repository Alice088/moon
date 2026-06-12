package hook

import (
	"context"

	"moon/internal/entity"
)

type Hook func(ctx context.Context, metrics *entity.Metrics) error
