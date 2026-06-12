package hook

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	_ "modernc.org/sqlite"

	"moon/internal/entity"
)

func WriteAlertToDB(dbPath string) Hook {
	var db *sql.DB
	var once sync.Once
	var initErr error

	return func(ctx context.Context, m *entity.Metrics) error {
		once.Do(func() {
			db, initErr = sql.Open("sqlite", dbPath)
			if initErr != nil {
				return
			}
			_, initErr = db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS alerts (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					type TEXT NOT NULL,
					message TEXT NOT NULL,
					data TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				)
			`)
		})
		if initErr != nil {
			return initErr
		}

		alert, ok := m.Get("alert").(entity.Alert)
		if !ok {
			return nil
		}

		var dataJSON []byte
		if alert.Data != nil {
			dataJSON, _ = json.Marshal(alert.Data)
		}

		_, err := db.ExecContext(ctx,
			"INSERT INTO alerts (type, message, data) VALUES (?, ?, ?)",
			alert.Type, alert.Message, string(dataJSON),
		)
		return err
	}
}
