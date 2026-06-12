package bot

import (
	"context"
	"log"
	"time"
)

func (b *Bot) Run(ctx context.Context) {
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("bot getUpdates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, upd := range updates {
			if upd.UpdateID >= offset {
				offset = upd.UpdateID + 1
			}
			b.handleMessage(ctx, upd.Message)
		}
	}
}
