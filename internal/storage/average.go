package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

func Average(dbPath string, alertType string, since time.Time) (types.Percent, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var avg types.Percent
	err = db.QueryRow(
		"SELECT COALESCE(AVG(json_extract(data, '$.usage')), 0) FROM alerts WHERE type = ? AND created_at >= ?",
		alertType, since.Format("2006-01-02 15:04:05"),
	).Scan(&avg)
	if err != nil {
		return 0, err
	}

	return avg, nil
}
