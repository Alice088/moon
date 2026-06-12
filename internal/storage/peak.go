package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

func Peak(dbPath string, alertType string, since time.Time) (types.Percent, error) {
	alertType = alertType + "_peak"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var peak types.Percent
	err = db.QueryRow(
		"SELECT COALESCE(MAX(json_extract(data, '$.usage')), 0) FROM alerts WHERE type = ? AND created_at >= ?",
		alertType, since.Format("2006-01-02 15:04:05"),
	).Scan(&peak)
	if err != nil {
		return 0, err
	}

	return peak, nil
}
