package trigger

import (
	"testing"
	"time"
)

func TestStale(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	longAgo := now.Add(-time.Hour)   // subscribed long before the message
	justNow := now.Add(-time.Second) // message arrives right after subscribe
	hue := func(at time.Time) []byte {
		return []byte(`{"button":1,"event":"short_release","last-updated":"` + at.Format(time.RFC3339Nano) + `"}`)
	}

	tests := []struct {
		name         string
		payload      []byte
		subscribedAt time.Time
		want         bool
	}{
		{"fresh press fires", hue(now.Add(-500 * time.Millisecond)), longAgo, false},
		{"fresh press right after subscribe fires", hue(now.Add(-500 * time.Millisecond)), justNow, false},
		{"slightly future timestamp (clock skew) fires", hue(now.Add(time.Second)), longAgo, false},
		{"old press is skipped", hue(now.Add(-3 * time.Hour)), justNow, true},
		{"old press after reconnect is skipped", hue(now.Add(-10 * time.Second)), longAgo, true},
		{"no timestamp right after subscribe is skipped", []byte(`{"action":"on"}`), justNow, true},
		{"no timestamp later fires", []byte(`{"action":"on"}`), longAgo, false},
		{"plain payload later fires", []byte(`ON`), longAgo, false},
		{"unparseable timestamp falls back to grace window", []byte(`{"last-updated":"yesterday"}`), justNow, true},
		{"unparseable timestamp later fires", []byte(`{"last-updated":"yesterday"}`), longAgo, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := Stale(tt.payload, tt.subscribedAt, now)
			if got != tt.want {
				t.Fatalf("Stale() = %v (%q), want %v", got, reason, tt.want)
			}
			if got && reason == "" {
				t.Fatal("stale without a reason")
			}
		})
	}
}
