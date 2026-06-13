package bot

import "context"

type Bot struct {
	token        string
	dbPath       string
	chatIDs      map[string]bool
	postJSONFunc func(ctx context.Context, method string, v any) error
	debug        bool
}

func New(token, dbPath string, allowedChatIDs ...string) *Bot {
	ids := make(map[string]bool)
	for _, id := range allowedChatIDs {
		ids[id] = true
	}
	return &Bot{token: token, dbPath: dbPath, chatIDs: ids, postJSONFunc: nil}
}

func (b *Bot) SetDebug(on bool) {
	b.debug = on
}
