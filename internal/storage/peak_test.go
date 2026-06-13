package storage

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"moon/pkg/types"
)

func seedTestDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "moon-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := sql.Open("sqlite", f.Name())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		os.Remove(f.Name())
		t.Fatal(err)
	}

	// Use UTC to match production: SQLite CURRENT_TIMESTAMP stores UTC.
	now := time.Now().UTC()
	for _, row := range []struct {
		typ, msg, data, ts string
	}{
		{"cpu_peak", "cpu high", `{"usage":85.0,"threshold":80}`, now.Add(-10 * time.Minute).Format("2006-01-02 15:04:05")},
		{"cpu_peak", "cpu high", `{"usage":92.3,"threshold":80}`, now.Add(-5 * time.Minute).Format("2006-01-02 15:04:05")},
		{"ram_peak", "ram high", `{"usage":76.0,"threshold":80}`, now.Add(-3 * time.Minute).Format("2006-01-02 15:04:05")},
		{"cpu_peak", "cpu high", `{"usage":88.7,"threshold":80}`, now.Add(-1 * time.Minute).Format("2006-01-02 15:04:05")},
		{"disk_peak", "disk high", `{"usage":91.0,"threshold":85}`, now.Add(-30 * time.Second).Format("2006-01-02 15:04:05")},
	} {
		_, err := db.Exec("INSERT INTO alerts (type, message, data, created_at) VALUES (?, ?, ?, ?)",
			row.typ, row.msg, row.data, row.ts)
		if err != nil {
			db.Close()
			os.Remove(f.Name())
			t.Fatal(err)
		}
	}
	db.Close()
	return f.Name()
}

func cleanupDB(t *testing.T, path string) {
	t.Helper()
	os.Remove(path)
}

func TestPeak(t *testing.T) {
	path := seedTestDB(t)
	defer cleanupDB(t, path)

	since := time.Now().Add(-1 * time.Hour)

	peak, err := Peak(path, "cpu", since)
	if err != nil {
		t.Fatal(err)
	}
	if peak != types.Percent(92.3) {
		t.Fatalf("expected 92.3, got %f", peak)
	}

	peak, err = Peak(path, "ram", since)
	if err != nil {
		t.Fatal(err)
	}
	if peak != types.Percent(76.0) {
		t.Fatalf("expected 76.0, got %f", peak)
	}

	peak, err = Peak(path, "disk", since)
	if err != nil {
		t.Fatal(err)
	}
	if peak != types.Percent(91.0) {
		t.Fatalf("expected 91.0, got %f", peak)
	}
}

func TestPeakNoResults(t *testing.T) {
	path := seedTestDB(t)
	defer cleanupDB(t, path)

	since := time.Now().Add(1 * time.Hour)

	peak, err := Peak(path, "cpu", since)
	if err != nil {
		t.Fatal(err)
	}
	if peak != types.Percent(0) {
		t.Fatalf("expected 0, got %f", peak)
	}
}
