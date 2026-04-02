package timeutil

import (
	"testing"
	"time"
)

func TestNowUTCUsesUTC(t *testing.T) {
	got := NowUTC()
	if got.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", got.Location())
	}
}

func TestNormalizeUTCConvertsTimeToUTC(t *testing.T) {
	input := time.Date(2026, 4, 2, 12, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))

	got := NormalizeUTC(input)
	want := time.Date(2026, 4, 2, 9, 30, 0, 0, time.UTC)

	if got.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", got.Location())
	}
	if !got.Equal(want) {
		t.Fatalf("NormalizeUTC() = %v, want %v", got, want)
	}
}
