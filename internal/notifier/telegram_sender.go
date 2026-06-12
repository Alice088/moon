package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"moon/internal/entity"
)

type telegramSender struct {
	token  string
	chatID string
}

func NewTelegramSender(botToken string, chatID string) entity.Notifier {
	return &telegramSender{token: botToken, chatID: chatID}
}

func (t *telegramSender) Send(ctx context.Context, alert entity.Alert) error {
	body := map[string]string{
		"chat_id":                  t.chatID,
		"text":                     t.formatMessage(alert),
		"parse_mode":               "MarkdownV2",
		"disable_web_page_preview": "true",
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.telegram.org/bot"+t.token+"/sendMessage",
		bytes.NewReader(b),
	)
	if err != nil {
		return fmt.Errorf("tg request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("tg send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tg status: %s", resp.Status)
	}

	return nil
}

func (t *telegramSender) formatMessage(alert entity.Alert) string {
	now := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("*Alert:* %s\n*Time:* %s\n*Message:* %s",
		t.escape(alert.Type), t.escape(now), t.escape(alert.Message))

	if data, ok := alert.Data.(map[string]any); ok {
		if u, ok := data["usage"]; ok {
			msg += fmt.Sprintf("\n*Usage:* %s%%", t.escape(fmt.Sprintf("%.1f", u)))
		}
		if th, ok := data["threshold"]; ok {
			msg += fmt.Sprintf("\n*Threshold:* %s%%", t.escape(fmt.Sprintf("%.1f", th)))
		}
	}

	return msg
}

func (t *telegramSender) escape(s string) string {
	var out []byte
	for _, r := range s {
		switch r {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			out = append(out, '\\', byte(r))
		default:
			out = append(out, byte(r))
		}
	}
	return string(out)
}
