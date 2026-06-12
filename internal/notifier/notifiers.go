package notifier

import (
	"moon/internal/config"
	"moon/internal/entity"
)

func NewNotifiers(cfgs []config.NotifyConfig) []entity.Notifier {
	var notifiers []entity.Notifier
	for _, c := range cfgs {
		switch c.Type {
		case "telegram":
			notifiers = append(notifiers, NewTelegramSender(c.BotToken, c.ChatID))
		case "mail":
			// TODO: NewMailSender(...)
		}
	}
	return notifiers
}
