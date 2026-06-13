package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

func Average(dbPath string, alertType string, since time.Time) (types.Percent, error) {
	alertType = alertType + "_peak"
	sinceStr := since.UTC().Format("2006-01-02 15:04:05")

	if Debug {
		log.Printf("[debug] storage.Average: db=%s type=%s since=%s", dbPath, alertType, sinceStr)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		if Debug {
			log.Printf("[debug] storage.Average: open db error: %v", err)
		}
		return 0, err
	}
	defer db.Close()

	var avg types.Percent
	err = db.QueryRow(
		"SELECT COALESCE(AVG(json_extract(data, '$.usage')), 0) FROM alerts WHERE type = ? AND created_at >= ?",
		alertType, sinceStr,
	).Scan(&avg)
	if err != nil {
		if Debug {
			log.Printf("[debug] storage.Average: query error: %v", err)
		}
		return 0, err
	}

	if Debug {
		log.Printf("[debug] storage.Average: result=%.1f%%", avg)
	}

	return avg, nil
}
