package bot

import (
	"testing"
	"time"
)

func eq(t *testing.T, got, want time.Duration) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParsePeriod(t *testing.T) {
	d, err := parsePeriod("hour")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, time.Hour)

	d, err = parsePeriod("day")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 24*time.Hour)

	d, err = parsePeriod("week")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 7*24*time.Hour)

	d, err = parsePeriod("month")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 30*24*time.Hour)

	d, err = parsePeriod("1h")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, time.Hour)

	d, err = parsePeriod("24h")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 24*time.Hour)

	d, err = parsePeriod("7d")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 7*24*time.Hour)

	d, err = parsePeriod("30d")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, d, 30*24*time.Hour)

	_, err = parsePeriod("invalid")
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}
