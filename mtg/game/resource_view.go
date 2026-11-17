package game

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// PermanentResources is a planned resource state, not a zone-change operation.
type PermanentResources struct {
	Removed  bool
	Tapped   bool
	Counters map[counter.Kind]int
}

// PlayerResources is a planned payment-resource state.
type PlayerResources struct {
	Life   int
	Energy int
	Mana   map[mana.Unit]int
	Riders []ManaRiderInstance
}

// ResourceCardMove records a deterministic payment-cost move in a read view.
type ResourceCardMove struct {
	Player PlayerID
	Card   id.ID
	From   zone.Type
	To     zone.Type
}

// ResourceProjection contains only payment resources and recorded ability uses.
// Missing entries retain their current state. It never allocates object IDs or
// changes RNG, logs or the underlying game. CardMoves changes only view zones.
type ResourceProjection struct {
	Permanents       map[id.ID]PermanentResources
	Players          map[PlayerID]PlayerResources
	Used             map[ActivatedAbilityUse]bool
	Events           []Event
	UnavailableCards map[id.ID]bool
	LastKnown        map[id.ID]ObjectSnapshot
	SpentDelayed     map[id.ID]bool
	CardMoves        []ResourceCardMove
}

// ResourceView supplies canonical rules evaluators with a read-only,
// copy-on-write resource projection. Unlike Clone it does not copy the game
// graph or fork randomness. Only affected resource records and the battlefield
// index are independent; every other shared record must be read only.
func (g *Game) ResourceView(projection ResourceProjection) *Game {
	view := *g
	view.staticFrame = nil
	view.computingCharacteristics = nil
	view.choiceCtx = nil
	view.pendingEntryEvents = nil
	view.ManaPaymentScopes = nil
	view.HiddenZoneSearchScopes = nil
	view.RevealedHands = nil
	view.SearchControlScopes = nil
	view.LibraryTopWindows = nil
	view.Battlefield = make([]*Permanent, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		resources, changed := projection.Permanents[permanent.ObjectID]
		if changed && resources.Removed {
			continue
		}
		if !changed {
			view.Battlefield = append(view.Battlefield, permanent)
			continue
		}
		projected := *permanent
		projected.Tapped = resources.Tapped
		projected.Counters = counter.Set{}
		for kind, amount := range resources.Counters {
			projected.Counters.Add(kind, amount)
		}
		projected.CounterTimestamps = maps.Clone(permanent.CounterTimestamps)
		view.Battlefield = append(view.Battlefield, &projected)
	}
	for playerID, resources := range projection.Players {
		if playerID < 0 || int(playerID) >= len(g.Players) || g.Players[playerID] == nil {
			continue
		}
		player := *g.Players[playerID]
		player.Life = resources.Life
		player.EnergyCounters = resources.Energy
		player.ManaPool = mana.NewPool()
		for unit, amount := range resources.Mana {
			player.ManaPool.AddUnit(unit, amount)
		}
		player.ManaRiders = slices.Clone(resources.Riders)
		view.Players[playerID] = &player
	}
	if len(projection.UnavailableCards) > 0 {
		for playerID, original := range view.Players {
			if original == nil {
				continue
			}
			player := *original
			player.Hand = original.Hand.Clone()
			for cardID := range projection.UnavailableCards {
				player.Hand.Remove(cardID)
			}
			view.Players[playerID] = &player
		}
	}
	if len(projection.CardMoves) > 0 {
		view.CardInstances = maps.Clone(g.CardInstances)
		for _, move := range projection.CardMoves {
			original := view.Players[move.Player]
			player := *original
			player.Hand, player.Graveyard, player.Exile = original.Hand.Clone(), original.Graveyard.Clone(), original.Exile.Clone()
			cards := func(kind zone.Type) *zone.Zone {
				switch kind {
				case zone.Hand:
					return &player.Hand
				case zone.Graveyard:
					return &player.Graveyard
				case zone.Exile:
					return &player.Exile
				default:
					panic("unsupported payment resource zone")
				}
			}
			cards(move.From).Remove(move.Card)
			cards(move.To).Add(move.Card)
			view.Players[move.Player] = &player
			card := *view.CardInstances[move.Card]
			card.ZoneVersion++
			view.CardInstances[move.Card] = &card
		}
	}
	if len(projection.Used) > 0 {
		view.ActivatedAbilitiesThisTurn = maps.Clone(g.ActivatedAbilitiesThisTurn)
		maps.Copy(view.ActivatedAbilitiesThisTurn, projection.Used)
	}
	if len(projection.LastKnown) > 0 {
		view.LastKnownInformation = maps.Clone(g.LastKnownInformation)
		maps.Copy(view.LastKnownInformation, projection.LastKnown)
	}
	if len(projection.SpentDelayed) > 0 {
		view.DelayedTriggers = make([]DelayedTrigger, 0, len(g.DelayedTriggers))
		for _, trigger := range g.DelayedTriggers {
			if !projection.SpentDelayed[trigger.ID] {
				view.DelayedTriggers = append(view.DelayedTriggers, trigger)
			}
		}
	}
	if len(projection.Events) > 0 {
		view.Events = append(slices.Clone(g.Events), projection.Events...)
	}
	return &view
}
