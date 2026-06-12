package storage

import (
	"testing"
	"time"

	"moon/pkg/types"
)

func TestAverage(t *testing.T) {
	path := seedTestDB(t)
	defer cleanupDB(t, path)

	since := time.Now().Add(-1 * time.Hour)

	avg, err := Average(path, "cpu", since)
	if err != nil {
		t.Fatal(err)
	}
	expected := types.Percent((85.0 + 92.3 + 88.7) / 3)
	if avg != expected {
		t.Fatalf("expected %f, got %f", expected, avg)
	}

	avg, err = Average(path, "ram", since)
	if err != nil {
		t.Fatal(err)
	}
	if avg != types.Percent(76.0) {
		t.Fatalf("expected 76.0, got %f", avg)
	}

	avg, err = Average(path, "disk", since)
	if err != nil {
		t.Fatal(err)
	}
	if avg != types.Percent(91.0) {
		t.Fatalf("expected 91.0, got %f", avg)
	}
}

func TestAverageNoResults(t *testing.T) {
	path := seedTestDB(t)
	defer cleanupDB(t, path)

	since := time.Now().Add(1 * time.Hour)

	avg, err := Average(path, "cpu", since)
	if err != nil {
		t.Fatal(err)
	}
	if avg != types.Percent(0) {
		t.Fatalf("expected 0, got %f", avg)
	}
}
