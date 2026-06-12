package notifier

import (
	"context"
	"fmt"

	"moon/internal/entity"
)

func NewTelegramSender(botToken string) entity.Notifier {
	return &telegramSender{token: botToken}
}

type telegramSender struct {
	token string
}

func (t *telegramSender) Send(ctx context.Context, alert entity.Alert) error {
	_ = t.token
	_ = alert
	return fmt.Errorf("telegram not implemented")
}
