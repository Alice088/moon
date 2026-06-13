package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

// AverageByIntervals divides [since, now] into n equal parts and returns the
// average value for each interval.
func AverageByIntervals(dbPath string, alertType string, since time.Time, n int) ([]IntervalPeak, error) {
	if n <= 0 {
		n = 3
	}

	alertType = alertType + "_peak"
	now := time.Now()
	duration := now.Sub(since)
	intervalDur := duration / time.Duration(n)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	results := make([]IntervalPeak, n)

	for i := 0; i < n; i++ {
		start := since.Add(intervalDur * time.Duration(i))
		end := since.Add(intervalDur * time.Duration(i + 1))

		results[i].Start = start
		results[i].End = end

		startStr := start.UTC().Format("2006-01-02 15:04:05")
		endStr := end.UTC().Format("2006-01-02 15:04:05")

		if Debug {
			log.Printf("[debug] storage.AverageByIntervals: db=%s type=%s interval=%d [%s, %s)",
				dbPath, alertType, i, startStr, endStr)
		}

		var avg types.Percent
		err := db.QueryRow(
			"SELECT COALESCE(AVG(json_extract(data, '$.usage')), 0) FROM alerts WHERE type = ? AND created_at >= ? AND created_at < ?",
			alertType, startStr, endStr,
		).Scan(&avg)
		if err != nil {
			if Debug {
				log.Printf("[debug] storage.AverageByIntervals: interval=%d query error: %v", i, err)
			}
			return nil, err
		}

		results[i].Value = avg

		if Debug {
			log.Printf("[debug] storage.AverageByIntervals: interval=%d result=%.1f%%", i, avg)
		}
	}

	return results, nil
}

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
