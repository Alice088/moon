package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"

	"moon/internal/entity"
)

func ListAlerts(dbPath string, alertType string, since time.Time) ([]entity.Alert, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT type, message, data FROM alerts WHERE type = ? AND created_at >= ? ORDER BY created_at DESC",
		alertType, since.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []entity.Alert
	for rows.Next() {
		var a entity.Alert
		var dataStr string
		if err := rows.Scan(&a.Type, &a.Message, &dataStr); err != nil {
			return nil, err
		}
		if dataStr != "" {
			json.Unmarshal([]byte(dataStr), &a.Data)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
