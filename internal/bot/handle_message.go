package bot

import (
	"context"
	"fmt"
	"log"
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
		if b.debug {
			log.Printf("[debug] handleMessage: bad period %q: %v", periodStr, err)
		}
		b.postJSON(ctx, "sendMessage", map[string]string{
			"chat_id":                  chatID,
			"text":                     "invalid period. use: hour, day, week, month, or Go duration (1h, 24h, 7d)",
			"disable_web_page_preview": "true",
		})
		return
	}

	since := time.Now().Add(-d)

	if b.debug {
		log.Printf("[debug] handleMessage: cmd=%s period=%s since=%s since_utc=%s",
			cmd, periodStr, since.Format("2006-01-02 15:04:05"), since.UTC().Format("2006-01-02 15:04:05"))
	}

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
	now := time.Now()
	duration := now.Sub(since)

	// choose time format based on total period length
	showDate := duration >= 24*time.Hour

	peakTimeFmt := "15:04"
	if showDate {
		peakTimeFmt = "Jan _2 15:04"
	}

	n := 3
	var lines []string
	if showDate {
		lines = append(lines, fmt.Sprintf("Peaks %s — %s:", since.Format("Jan _2"), now.Format("Jan _2")))
	} else {
		lines = append(lines, fmt.Sprintf("Peaks %s — %s:", since.Format("15:04"), now.Format("15:04")))
	}

	for _, t := range types {
		intervals, err := storage.PeakByIntervals(b.dbPath, t, since, n)
		if err != nil {
			if b.debug {
				log.Printf("[debug] sendPeaks %s: storage error: %v", t, err)
			}
			lines = append(lines, fmt.Sprintf("  %s: error", t))
			continue
		}

		lines = append(lines, fmt.Sprintf("  %s:", t))
		for _, iv := range intervals {
			if iv.PeakAt == nil {
				lines = append(lines, fmt.Sprintf("    no data"))
			} else {
				lines = append(lines, fmt.Sprintf("    %s: %.1f%%", iv.PeakAt.Local().Format(peakTimeFmt), iv.Value))
			}
		}
	}

	text := strings.Join(lines, "\n")
	if b.debug {
		log.Printf("[debug] sendPeaks text:\n%s", text)
	}
	b.postJSON(ctx, "sendMessage", map[string]string{
		"chat_id":                  chatID,
		"text":                     text,
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
			if b.debug {
				log.Printf("[debug] sendPeakAvg %s: storage error: %v", t, err)
			}
			a = 0
		}
		if b.debug {
			log.Printf("[debug] sendPeakAvg %s: %.1f%%", t, a)
		}
		lines = append(lines, fmt.Sprintf("  %s: %.1f%%", t, a))
	}
	if b.debug {
		log.Printf("[debug] sendPeakAvg text:\n%s", strings.Join(lines, "\n"))
	}
	b.postJSON(ctx, "sendMessage", map[string]string{
		"chat_id":                  chatID,
		"text":                     strings.Join(lines, "\n"),
		"disable_web_page_preview": "true",
	})
}
