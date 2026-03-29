package application

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// scheduleWriteGate serializes same-key schedule writes in-process using channels.
// Database constraints remain the source of truth for cross-process safety.
type scheduleWriteGate struct {
	mu      sync.Mutex
	entries map[string]*scheduleGateEntry
}

type scheduleGateEntry struct {
	ch   chan struct{}
	refs int
}

func newScheduleWriteGate() *scheduleWriteGate {
	return &scheduleWriteGate{entries: make(map[string]*scheduleGateEntry)}
}

func (g *scheduleWriteGate) WithKey(ctx context.Context, key string, fn func() error) error {
	entry := g.acquireEntry(key)

	select {
	case <-ctx.Done():
		g.releaseEntry(key, entry)
		return ctx.Err()
	case <-entry.ch:
	}

	defer func() {
		entry.ch <- struct{}{}
		g.releaseEntry(key, entry)
	}()

	return fn()
}

func (g *scheduleWriteGate) acquireEntry(key string) *scheduleGateEntry {
	g.mu.Lock()
	defer g.mu.Unlock()

	entry, ok := g.entries[key]
	if !ok {
		entry = &scheduleGateEntry{ch: make(chan struct{}, 1)}
		entry.ch <- struct{}{}
		g.entries[key] = entry
	}
	entry.refs++
	return entry
}

func (g *scheduleWriteGate) releaseEntry(key string, entry *scheduleGateEntry) {
	g.mu.Lock()
	defer g.mu.Unlock()

	current, ok := g.entries[key]
	if !ok || current != entry {
		return
	}

	current.refs--
	if current.refs == 0 {
		delete(g.entries, key)
	}
}

func scheduleCreateGateKey(companyID, homeTeamID, guestTeamID int64, matchDate, matchTime time.Time) string {
	return fmt.Sprintf(
		"schedule:create:%d:%d:%d:%s:%s",
		companyID,
		homeTeamID,
		guestTeamID,
		matchDate.UTC().Format("2006-01-02"),
		matchTime.UTC().Format("15:04:05"),
	)
}
