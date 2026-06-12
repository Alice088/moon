package bot

type Bot struct {
	token   string
	dbPath  string
	chatIDs map[string]bool
}

func New(token, dbPath string, allowedChatIDs ...string) *Bot {
	ids := make(map[string]bool)
	for _, id := range allowedChatIDs {
		ids[id] = true
	}
	return &Bot{token: token, dbPath: dbPath, chatIDs: ids}
}
