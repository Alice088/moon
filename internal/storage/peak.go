package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

// IntervalPeak holds the max value and its timestamp for one time interval.
type IntervalPeak struct {
	Start  time.Time
	End    time.Time
	Value  types.Percent
	PeakAt *time.Time // nil if no alerts in this interval
}

// Debug enables verbose logging in storage functions.
var Debug bool

func Peak(dbPath string, alertType string, since time.Time) (types.Percent, error) {
	alertType = alertType + "_peak"
	sinceStr := since.UTC().Format("2006-01-02 15:04:05")

	if Debug {
		log.Printf("[debug] storage.Peak: db=%s type=%s since=%s", dbPath, alertType, sinceStr)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		if Debug {
			log.Printf("[debug] storage.Peak: open db error: %v", err)
		}
		return 0, err
	}
	defer db.Close()

	var peak types.Percent
	err = db.QueryRow(
		"SELECT COALESCE(MAX(json_extract(data, '$.usage')), 0) FROM alerts WHERE type = ? AND created_at >= ?",
		alertType, sinceStr,
	).Scan(&peak)
	if err != nil {
		if Debug {
			log.Printf("[debug] storage.Peak: query error: %v", err)
		}
		return 0, err
	}

	if Debug {
		log.Printf("[debug] storage.Peak: result=%.1f%%", peak)
	}

	return peak, nil
}

// PeakByIntervals divides [since, now] into n equal parts and returns the
// highest peak (with timestamp) for each interval.
func PeakByIntervals(dbPath string, alertType string, since time.Time, n int) ([]IntervalPeak, error) {
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
			log.Printf("[debug] storage.PeakByIntervals: db=%s type=%s interval=%d [%s, %s)",
				dbPath, alertType, i, startStr, endStr)
		}

		var usage types.Percent
		var createdAt sql.NullString
		err := db.QueryRow(
			`SELECT json_extract(data, '$.usage'), created_at
			 FROM alerts
			 WHERE type = ? AND created_at >= ? AND created_at < ?
			 ORDER BY json_extract(data, '$.usage') DESC, created_at ASC
			 LIMIT 1`,
			alertType, startStr, endStr,
		).Scan(&usage, &createdAt)

		if err == sql.ErrNoRows {
			results[i].Value = 0
			if Debug {
				log.Printf("[debug] storage.PeakByIntervals: interval=%d no data", i)
			}
		} else if err != nil {
			if Debug {
				log.Printf("[debug] storage.PeakByIntervals: interval=%d query error: %v", i, err)
			}
			return nil, err
		} else {
			results[i].Value = usage
			if createdAt.Valid {
				t, parseErr := time.Parse("2006-01-02 15:04:05", createdAt.String)
				if parseErr != nil {
					// maybe SQLite returned a different format
					for _, fmt := range []string{
						"2006-01-02 15:04:05",
						"2006-01-02T15:04:05Z",
						"2006-01-02T15:04:05.000Z",
						"2006-01-02 15:04:05.000",
						"2006-01-02T15:04:05-07:00",
						"2006-01-02 15:04:05-07:00",
					} {
						t, err2 := time.Parse(fmt, createdAt.String)
						if err2 == nil {
							results[i].PeakAt = &t
							break
						}
					}
				} else {
					results[i].PeakAt = &t
				}
			}
			if Debug {
				local := "(none)"
				if results[i].PeakAt != nil {
					local = results[i].PeakAt.Format(time.RFC3339)
				} else {
					local = fmt.Sprintf("(parse failed: raw=%q)", createdAt.String)
				}
				log.Printf("[debug] storage.PeakByIntervals: interval=%d value=%.1f%% at %s", i, usage, local)
			}
		}
	}

	return results, nil
}
