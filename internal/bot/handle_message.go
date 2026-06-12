package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"moon/internal/storage"
)

type tgUpdate struct {
	UpdateID int       `json:"update_id"`
	Message  tgMessage `json:"message"`
}

type tgMessage struct {
	MessageID int    `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

func (b *Bot) handleMessage(ctx context.Context, msg tgMessage) {
	chatID := fmt.Sprintf("%d", msg.Chat.ID)
	if !b.chatIDs[chatID] {
		return
	}

	text := strings.TrimSpace(msg.Text)
	if !strings.HasPrefix(text, "/") {
		return
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		return
	}

	cmd := parts[0]
	periodStr := parts[1]

	d, err := parsePeriod(periodStr)
	if err != nil {
		b.postJSON(ctx, "sendMessage", map[string]string{
			"chat_id":                  chatID,
			"text":                     "invalid period. use: hour, day, week, month, or Go duration (1h, 24h, 7d)",
			"disable_web_page_preview": "true",
		})
		return
	}

	since := time.Now().Add(-d)

	switch cmd {
	case "/peaks":
		b.sendPeaks(ctx, chatID, since)
	case "/peak-avg":
		b.sendPeakAvg(ctx, chatID, since)
	default:
		b.postJSON(ctx, "sendMessage", map[string]string{
			"chat_id":                  chatID,
			"text":                     "unknown command. use /peaks <period> or /peak-avg <period>",
			"disable_web_page_preview": "true",
		})
	}
}

func (b *Bot) sendPeaks(ctx context.Context, chatID string, since time.Time) {
	types := []string{"cpu", "ram", "disk"}
	var lines []string
	lines = append(lines, fmt.Sprintf("Peaks since %s:", since.Format("2006-01-02 15:04")))
	for _, t := range types {
		p, err := storage.Peak(b.dbPath, t, since)
		if err != nil {
			p = 0
		}
		lines = append(lines, fmt.Sprintf("  %s: %.1f%%", t, p))
	}
	b.postJSON(ctx, "sendMessage", map[string]string{
		"chat_id":                  chatID,
		"text":                     strings.Join(lines, "\n"),
		"disable_web_page_preview": "true",
	})
}

func (b *Bot) sendPeakAvg(ctx context.Context, chatID string, since time.Time) {
	types := []string{"cpu", "ram", "disk"}
	var lines []string
	lines = append(lines, fmt.Sprintf("Average peaks since %s:", since.Format("2006-01-02 15:04")))
	for _, t := range types {
		a, err := storage.Average(b.dbPath, t, since)
		if err != nil {
			a = 0
		}
		lines = append(lines, fmt.Sprintf("  %s: %.1f%%", t, a))
	}
	b.postJSON(ctx, "sendMessage", map[string]string{
		"chat_id":                  chatID,
		"text":                     strings.Join(lines, "\n"),
		"disable_web_page_preview": "true",
	})
}
