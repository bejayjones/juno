// Package domain defines the core types for the sync bounded context.
package domain

import (
	"sync/atomic"
	"time"
)

// SyncRecord mirrors a row in the sync_records table. Each row represents
// one mutation that needs to be (or has been) pushed to the cloud.
type SyncRecord struct {
	ID           string
	TableName    string
	RecordID     string
	Operation    string // "upsert" | "delete"
	Payload      string // JSON snapshot of the affected row
	LamportClock int64
	Synced       bool
	CreatedAt    time.Time
}

// LamportClock is a monotonically increasing logical timestamp used for
// last-writer-wins conflict resolution across devices.
//
// Rules:
//   - Each local write calls Tick(), which returns the next clock value.
//   - When receiving remote records, call Witness(remoteClock) to advance
//     the local clock past the remote value.
//   - Conflicts are resolved by keeping the record with the higher clock;
//     findings are append-only (never deleted during sync).
type LamportClock struct {
	v atomic.Int64
}

// NewLamportClock creates a clock initialised to initialValue.
// Pass the maximum lamport_clock value from the local sync_records table
// on startup so the clock survives restarts.
func NewLamportClock(initialValue int64) *LamportClock {
	c := &LamportClock{}
	c.v.Store(initialValue)
	return c
}

// Tick increments the clock by one and returns the new value.
func (c *LamportClock) Tick() int64 {
	return c.v.Add(1)
}

// Witness advances the clock to max(current, remote) + 1, ensuring that
// any subsequent local event has a strictly higher clock than the remote one.
func (c *LamportClock) Witness(remote int64) {
	for {
		cur := c.v.Load()
		var next int64
		if remote > cur {
			next = remote + 1
		} else {
			next = cur + 1
		}
		if c.v.CompareAndSwap(cur, next) {
			return
		}
	}
}

// Current returns the current clock value without advancing it.
func (c *LamportClock) Current() int64 {
	return c.v.Load()
}
