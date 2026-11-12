package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exileCardCounterChange puts counters on or removes counters from cards in
// exile on behalf of one player as one simultaneous action, then publishes the
// matching EventCountersAdded and EventCountersRemoved events once every change
// has happened. Counters on exiled cards live in Game.CardCounters, the
// canonical off-battlefield store. Each event identifies the exiled card by
// CardID and CardZoneVersion (with no PermanentID), and records the counter
// kind, how many of that kind the card had immediately before, and how many
// were actually put on or removed.
type exileCardCounterChange struct {
	g          *game.Game
	controller game.PlayerID
	events     []game.Event
}

func newExileCardCounterChange(g *game.Game, controller game.PlayerID) *exileCardCounterChange {
	return &exileCardCounterChange{g: g, controller: controller}
}

// exiledCard returns cardID's instance when it is in its owner's exile zone.
func (c *exileCardCounterChange) exiledCard(cardID id.ID) (*game.CardInstance, bool) {
	card, ok := c.g.GetCardInstance(cardID)
	if !ok {
		return nil, false
	}
	owner, ok := playerByID(c.g, card.Owner)
	if !ok || !owner.Exile.Contains(cardID) {
		return nil, false
	}
	return card, true
}

func (c *exileCardCounterChange) event(card *game.CardInstance, kind game.EventKind, counterKind counter.Kind, previous, amount int) game.Event {
	return game.Event{
		Kind:                  kind,
		Controller:            c.controller,
		Player:                card.Owner,
		CardID:                card.ID,
		CardZoneVersion:       card.ZoneVersion,
		CounterKind:           counterKind,
		PreviousCounterAmount: previous,
		Amount:                amount,
	}
}

// add puts amount counters of kind on the exiled card cardID and returns the
// number put on. A card that isn't in exile gets none.
func (c *exileCardCounterChange) add(cardID id.ID, kind counter.Kind, amount int) int {
	card, ok := c.exiledCard(cardID)
	if !ok || amount <= 0 {
		return 0
	}
	previous := c.g.ExileCounterCount(cardID, kind)
	c.g.AddExileCounter(cardID, kind, amount)
	c.events = append(c.events, c.event(card, game.EventCountersAdded, kind, previous, amount))
	return amount
}

// remove removes up to amount counters of kind from the exiled card cardID and
// returns the number removed.
func (c *exileCardCounterChange) remove(cardID id.ID, kind counter.Kind, amount int) int {
	card, ok := c.exiledCard(cardID)
	if !ok || amount <= 0 {
		return 0
	}
	set := c.g.CardCounters[cardID]
	previous := set.Get(kind)
	removed := set.Remove(kind, amount)
	if removed == 0 {
		return 0
	}
	if set.IsEmpty() {
		delete(c.g.CardCounters, cardID)
	} else {
		c.g.CardCounters[cardID] = set
	}
	c.events = append(c.events, c.event(card, game.EventCountersRemoved, kind, previous, removed))
	return removed
}

// publish emits the collected events. Changes to more than one card or of more
// than one kind happened simultaneously and share a SimultaneousID.
func (c *exileCardCounterChange) publish() {
	if len(c.events) > 1 {
		simultaneousID := c.g.IDGen.Next()
		for i := range c.events {
			c.events[i].SimultaneousID = simultaneousID
		}
	}
	for _, event := range c.events {
		emitEvent(c.g, event)
	}
	c.events = nil
}

// exileCardCounterEvent reports whether event put counters on or removed
// counters from a card off the battlefield rather than a permanent or player.
// Abilities that refer to counters being put on or removed from an object
// observe permanents (CR 122.6), so ordinary permanent-counter triggers ignore
// these events.
func exileCardCounterEvent(event game.Event) bool {
	return (event.Kind == game.EventCountersAdded || event.Kind == game.EventCountersRemoved) &&
		event.PermanentID == 0 && event.CardID != 0
}

// cardIsSuspended reports whether cardID is suspended: it is in exile, has
// suspend, and has a time counter on it (CR 702.62b). Suspend is checked among
// the abilities the face-up exiled card has there, including granted ones; a
// face-down exiled card has none. The abilities of a suspended card are
// controlled by its owner (CR 108.4a).
func cardIsSuspended(g *game.Game, cardID id.ID) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return g.HasExileCounter(cardID, counter.Time) && exiledCardHasSuspend(g, card)
}

// exiledCardHasSuspend reports whether card is in its owner's exile zone and
// has suspend among the abilities it has there, including granted ones. A
// face-down exiled card has none.
func exiledCardHasSuspend(g *game.Game, card *game.CardInstance) bool {
	owner, ok := playerByID(g, card.Owner)
	if !ok || !owner.Exile.Contains(card.ID) {
		return false
	}
	for _, ability := range cardZoneAbilities(g, card, zone.Exile, game.FaceFront, owner.Exile.IsFaceDown(card.ID)) {
		if _, ok := game.BodyKeywordAbility(ability, game.Suspend); ok {
			return true
		}
	}
	return false
}
