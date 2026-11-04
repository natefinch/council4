package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// siegeDefeatAbility is the intrinsic triggered ability every Siege has
// (CR 310.12b). It is shared immutable data; its behavior is
// resolveSiegeDefeat, selected by StackObject.SiegeDefeat.
var siegeDefeatAbility = &game.TriggeredAbility{
	Text: "When the last defense counter is removed from this permanent, exile it, then you may cast it transformed without paying its mana cost.",
}

func counterRemovedEvent(permanent *game.Permanent, controller game.PlayerID, kind counter.Kind, previous, amount int, simultaneousID id.ID) game.Event {
	return game.Event{
		Kind:                  game.EventCountersRemoved,
		SourceID:              permanent.CardInstanceID,
		SourceObjectID:        permanent.ObjectID,
		Controller:            controller,
		CardID:                permanent.CardInstanceID,
		PermanentID:           permanent.ObjectID,
		CounterKind:           kind,
		PreviousCounterAmount: previous,
		Amount:                amount,
		SimultaneousID:        simultaneousID,
	}
}

// counterRemoval removes counters from permanents on behalf of one player as
// one simultaneous action by an effect or cost, then publishes the matching
// EventCountersRemoved events once every removal has happened. Each event
// records the number of that kind actually removed and how many the permanent
// had immediately before, so a Siege whose last defense counter is removed
// triggers (CR 310.12b) however the counters are removed.
type counterRemoval struct {
	g          *game.Game
	controller game.PlayerID
	events     []game.Event
}

func newCounterRemoval(g *game.Game, controller game.PlayerID) *counterRemoval {
	return &counterRemoval{g: g, controller: controller}
}

// remove removes up to amount counters of kind from permanent and returns the
// number removed.
func (c *counterRemoval) remove(permanent *game.Permanent, kind counter.Kind, amount int) int {
	if permanent == nil || amount <= 0 {
		return 0
	}
	previous := permanent.Counters.Get(kind)
	removed := permanent.Counters.Remove(kind, amount)
	if removed > 0 {
		c.events = append(c.events, counterRemovedEvent(permanent, c.controller, kind, previous, removed, 0))
	}
	return removed
}

// removeAll removes every counter of every kind from permanent ("remove all
// counters", Vampire Hexmage) and returns the total removed. Kinds are removed
// in counter-kind order so the published events are reproducible.
func (c *counterRemoval) removeAll(permanent *game.Permanent) int {
	if permanent == nil {
		return 0
	}
	counts := permanent.Counters.All()
	total := 0
	for _, kind := range slices.Sorted(maps.Keys(counts)) {
		total += c.remove(permanent, kind, counts[kind])
	}
	return total
}

// publish emits the collected events. Removals from more than one permanent or
// of more than one kind happened simultaneously and share a SimultaneousID.
func (c *counterRemoval) publish() {
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

// removeCountersFromPermanent removes up to amount counters of kind from
// permanent on behalf of controller, publishes the removal and returns the
// number removed.
func removeCountersFromPermanent(g *game.Game, controller game.PlayerID, permanent *game.Permanent, kind counter.Kind, amount int) int {
	removal := newCounterRemoval(g, controller)
	removed := removal.remove(permanent, kind, amount)
	removal.publish()
	return removed
}

// siegeDefeatTriggeredByEvent reports whether event removed the last defense
// counter from permanent. A permanent that never had a defense counter can't
// have its last one removed, and putting a 0-defense battle into a graveyard
// removes no counter (Siege rulings 2023-04-14).
func siegeDefeatTriggeredByEvent(permanent *game.Permanent, event game.Event) bool {
	return event.Kind == game.EventCountersRemoved &&
		event.PermanentID == permanent.ObjectID &&
		event.CounterKind == counter.Defense &&
		event.PreviousCounterAmount > 0 &&
		event.Amount >= event.PreviousCounterAmount
}

// resolveSiegeDefeat resolves a Siege's intrinsic ability: exile the source
// permanent, then its controller may cast the exiled card transformed without
// paying its mana cost (CR 310.12b, 712.11a). Only the card the source became
// when exiled is offered; a token ceases to exist, and a card without a
// transforming back face can't be cast transformed and stays in exile. A
// source that already left the battlefield, or that a replacement put
// elsewhere, is not followed.
func (e *Engine) resolveSiegeDefeat(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	source, ok := permanentByObjectID(g, obj.SourceID)
	if !ok || source.PhasedOut {
		return "resolved"
	}
	for _, result := range movePermanentsToZoneSimultaneouslyWithResults(g, []*game.Permanent{source}, zone.Exile) {
		if !result.moved || result.destination != zone.Exile {
			continue
		}
		for _, arrival := range result.arrivals {
			card, ok := g.GetCardInstance(arrival.CardID)
			if !ok || card.ZoneVersion != arrival.ZoneVersion {
				continue
			}
			e.chooseAndCastFreeSpell(g, g.Players[card.Owner], obj.Controller, card.ID, zone.Exile,
				game.Selection{}, opt.V[int]{}, freeCastSpell{transformed: true}, true, agents, log)
		}
	}
	return "resolved"
}
