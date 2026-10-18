package main

import (
	"time"

	"github.com/tidwall/gjson"
)

// Trigger topics such as hue/button/<name> are retained, so every
// (re)subscribe hands us the last button press again. Acting on it would
// replay that press on each restart (e.g. switch the machine on unattended).
//
// paho's Message.Retained() would be the clean signal, but the mqtt-gateway
// callback only passes topic and payload, so staleness is decided from what
// we have:
//
//  1. Payloads carrying a "last-updated" timestamp (Hue buttons) are stale
//     when that timestamp is older than triggerMaxAge. This also covers
//     replays after a broker reconnect, at any time.
//  2. Payloads without a usable timestamp (non-Hue topics) are stale when
//     they arrive within triggerStartupGrace after subscribing, which is
//     exactly when the broker delivers retained messages. A real event in
//     those first seconds is lost, which is the safe side for a machine
//     that heats up. Replays after a later reconnect cannot be told apart
//     from real events for these topics.
const (
	// triggerMaxAge is how old a timestamped event may be and still fire.
	// mqtt-rules uses 2 s for wall switches; 5 s leaves room for clock skew
	// and bridge latency while still rejecting any retained replay.
	triggerMaxAge = 5 * time.Second

	// triggerStartupGrace is how long after subscribing timestamp-less
	// messages are treated as retained replays.
	triggerStartupGrace = 3 * time.Second
)

// staleTrigger reports whether a trigger message must be ignored and why.
func staleTrigger(payload []byte, subscribedAt, now time.Time) (bool, string) {
	if ts := gjson.GetBytes(payload, "last-updated"); ts.Exists() {
		if lastUpdated, err := time.Parse(time.RFC3339Nano, ts.String()); err == nil {
			if age := now.Sub(lastUpdated); age > triggerMaxAge {
				return true, "last-updated is " + age.Round(time.Second).String() + " old"
			}
			return false, ""
		}
	}

	if now.Sub(subscribedAt) < triggerStartupGrace {
		return true, "no last-updated and received right after subscribe (retained replay)"
	}
	return false, ""
}
