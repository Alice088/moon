package entity

import "context"

type Notifier interface {
	Send(ctx context.Context, alert Alert) error
}
