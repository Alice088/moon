package bot

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// seedPeaksDB creates a temp SQLite DB with alert data for testing.
// All created_at are stored as UTC strings (same as SQLite CURRENT_TIMESTAMP).
func seedPeaksDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "moon-bot-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := sql.Open("sqlite", f.Name())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}
	defer db.Close()

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
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for _, row := range []struct {
		typ, msg, data, ts string
	}{
		{"cpu_peak", "cpu high", `{"usage":85.0,"threshold":80}`, now.Add(-10 * time.Minute).Format("2006-01-02 15:04:05")},
		{"cpu_peak", "cpu high", `{"usage":92.3,"threshold":80}`, now.Add(-5 * time.Minute).Format("2006-01-02 15:04:05")},
		{"cpu_peak", "cpu high", `{"usage":88.7,"threshold":80}`, now.Add(-1 * time.Minute).Format("2006-01-02 15:04:05")},
		{"ram_peak", "ram high", `{"usage":76.0,"threshold":80}`, now.Add(-3 * time.Minute).Format("2006-01-02 15:04:05")},
		{"disk_peak", "disk high", `{"usage":91.0,"threshold":85}`, now.Add(-30 * time.Second).Format("2006-01-02 15:04:05")},
	} {
		_, err := db.Exec("INSERT INTO alerts (type, message, data, created_at) VALUES (?, ?, ?, ?)",
			row.typ, row.msg, row.data, row.ts)
		if err != nil {
			t.Fatal(err)
		}
	}

	return f.Name()
}

func cleanupDB(t *testing.T, path string) {
	t.Helper()
	os.Remove(path)
}

// mockPostJSON returns a function that captures the method and payload for later inspection.
func mockPostJSON(t *testing.T, captured *struct {
	method string
	text   string
}) func(ctx context.Context, method string, v any) error {
	t.Helper()
	return func(ctx context.Context, method string, v any) error {
		captured.method = method
		m, ok := v.(map[string]string)
		if !ok {
			t.Fatalf("expected map[string]string, got %T", v)
		}
		captured.text = m["text"]
		return nil
	}
}

func TestSendPeaks(t *testing.T) {
	path := seedPeaksDB(t)
	defer cleanupDB(t, path)

	var captured struct {
		method string
		text   string
	}

	bot := &Bot{
		dbPath:       path,
		chatIDs:      map[string]bool{"123": true},
		postJSONFunc: mockPostJSON(t, &captured),
	}

	since := time.Now().Add(-1 * time.Hour)
	bot.sendPeaks(context.Background(), "123", since)

	if captured.method != "sendMessage" {
		t.Fatalf("expected method sendMessage, got %s", captured.method)
	}

	// Check header
	if !strings.Contains(captured.text, "Peaks") {
		t.Fatalf("missing header in: %s", captured.text)
	}

	// cpu peak (92.3 is max, should appear in last interval)
	if !strings.Contains(captured.text, "92.3%") {
		t.Fatalf("expected 92.3%% in interval, got:\n%s", captured.text)
	}
	// ram peak (76.0)
	if !strings.Contains(captured.text, "76.0%") {
		t.Fatalf("expected 76.0%% in interval, got:\n%s", captured.text)
	}
	// disk peak (91.0)
	if !strings.Contains(captured.text, "91.0%") {
		t.Fatalf("expected 91.0%% in interval, got:\n%s", captured.text)
	}

	// check no pipe separators
	if strings.Contains(captured.text, "|") {
		t.Fatalf("expected no | pipes, got:\n%s", captured.text)
	}

	// check metric names on their own lines
	if !strings.Contains(captured.text, "\n  cpu:\n") {
		t.Fatalf("expected cpu on its own line, got:\n%s", captured.text)
	}
}

func TestSendPeakAvg(t *testing.T) {
	path := seedPeaksDB(t)
	defer cleanupDB(t, path)

	var captured struct {
		method string
		text   string
	}

	bot := &Bot{
		dbPath:       path,
		chatIDs:      map[string]bool{"123": true},
		postJSONFunc: mockPostJSON(t, &captured),
	}

	since := time.Now().Add(-1 * time.Hour)
	bot.sendPeakAvg(context.Background(), "123", since)

	if captured.method != "sendMessage" {
		t.Fatalf("expected method sendMessage, got %s", captured.method)
	}

	if !strings.Contains(captured.text, "Average peaks since") {
		t.Fatalf("missing header in: %s", captured.text)
	}

	// cpu average: (85.0 + 92.3 + 88.7) / 3 = 88.666... → 88.7
	expectedCPU := "cpu: 88.7%"
	if !strings.Contains(captured.text, expectedCPU) {
		t.Fatalf("expected %s, got:\n%s", expectedCPU, captured.text)
	}
	if !strings.Contains(captured.text, "ram: 76.0%") {
		t.Fatalf("expected ram: 76.0%%, got:\n%s", captured.text)
	}
	if !strings.Contains(captured.text, "disk: 91.0%") {
		t.Fatalf("expected disk: 91.0%%, got:\n%s", captured.text)
	}
}

func TestSendPeaksNoData(t *testing.T) {
	path := seedPeaksDB(t)
	defer cleanupDB(t, path)

	var captured struct {
		method string
		text   string
	}

	bot := &Bot{
		dbPath:       path,
		chatIDs:      map[string]bool{"123": true},
		postJSONFunc: mockPostJSON(t, &captured),
	}

	// Use a time in the future — no alerts match → all "no data"
	since := time.Now().Add(1 * time.Hour)
	bot.sendPeaks(context.Background(), "123", since)

	if !strings.Contains(captured.text, "no data") {
		t.Fatalf("expected 'no data' for no data, got:\n%s", captured.text)
	}
}

// TestSendPeaksTimezoneBug verifies that since.UTC() is used in the SQL query.
// SQLite stores created_at as UTC. If `since` is formatted in local timezone,
// the comparison `created_at >= since` may exclude all rows when local time
// is ahead of UTC (e.g., MSK +3).
func TestSendPeaksTimezoneBug(t *testing.T) {
	path := seedPeaksDB(t)
	defer cleanupDB(t, path)

	var captured struct {
		method string
		text   string
	}

	bot := &Bot{
		dbPath:       path,
		chatIDs:      map[string]bool{"123": true},
		postJSONFunc: mockPostJSON(t, &captured),
	}

	// MSK timezone (UTC+3) — without .UTC() in storage, since would be 3h ahead
	// of DB timestamps, causing all rows to be excluded → 0.0%.
	loc := time.FixedZone("MSK", 3*60*60)
	since := time.Now().In(loc).Add(-1 * time.Hour)

	bot.sendPeaks(context.Background(), "123", since)

	// Should still find data despite MSK timezone
	if !strings.Contains(captured.text, "92.3%") {
		t.Fatalf("expected cpu 92.3%% with MSK timezone, got:\n%s", captured.text)
	}
}

func TestSendPeakAvgNoData(t *testing.T) {
	path := seedPeaksDB(t)
	defer cleanupDB(t, path)

	var captured struct {
		method string
		text   string
	}

	bot := &Bot{
		dbPath:       path,
		chatIDs:      map[string]bool{"123": true},
		postJSONFunc: mockPostJSON(t, &captured),
	}

	since := time.Now().Add(1 * time.Hour)
	bot.sendPeakAvg(context.Background(), "123", since)

	if !strings.Contains(captured.text, "cpu: 0.0%") {
		t.Fatalf("expected 0.0%% for no data, got:\n%s", captured.text)
	}
}
