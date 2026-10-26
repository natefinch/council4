package game

// PendingEntryEvents returns battlefield-entry events that have already
// happened but are still being prepared for the event log. The rules layer
// installs them only while it snapshots an entering permanent's characteristics,
// so current-turn history conditions (such as Celebration's "two or more nonland
// permanents entered the battlefield under your control this turn") see the
// entry itself and any simultaneous entries (CR 603.6a, 603.10).
func (g *Game) PendingEntryEvents() []Event {
	return g.pendingEntryEvents
}

// SetPendingEntryEvents installs the transient pending-entry overlay and
// returns the previous one so the caller can restore it. The overlay is never
// deep copied (Clone starts without it).
func (g *Game) SetPendingEntryEvents(events []Event) []Event {
	previous := g.pendingEntryEvents
	g.pendingEntryEvents = events
	return previous
}
