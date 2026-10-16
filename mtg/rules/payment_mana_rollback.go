package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func reversibleManaCheckpoint(g *game.Game, log *TurnLog) func() {
	before := g.Clone()
	players, cards := g.Players, maps.Clone(g.CardInstances)
	permanents := make(map[id.ID]*game.Permanent, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		permanents[permanent.ObjectID] = permanent
	}
	objects := make(map[id.ID]*game.StackObject)
	for _, object := range g.Stack.Objects() {
		objects[object.ID] = object
	}
	entries := 0
	if log != nil {
		entries = len(log.Entries)
	}
	return func() {
		// Keep synchronous Engine references alive, including a search's Player
		// and a proposed spell's StackObject. Clone's RNG is not a continuation;
		// reversible activations did not consume the live random stream.
		for player, saved := range before.Players {
			if saved != nil && players[player] != nil {
				*players[player] = *saved
				before.Players[player] = players[player]
			}
		}
		for cardID, saved := range before.CardInstances {
			if live := cards[cardID]; live != nil {
				*live = *saved
				before.CardInstances[cardID] = live
			}
		}
		for index, saved := range before.Battlefield {
			if live := permanents[saved.ObjectID]; live != nil {
				*live = *saved
				before.Battlefield[index] = live
			}
		}
		var stack game.Stack
		for _, saved := range before.Stack.Objects() {
			if live := objects[saved.ID]; live != nil {
				*live = *saved
				stack.Push(live)
			} else {
				stack.Push(saved)
			}
		}
		before.Stack = stack
		g.RestoreReversibleManaPayment(before)
		if log != nil {
			// Choices remain recorded so failed proposals replay identically.
			log.Entries = log.Entries[:entries]
		}
	}
}

func manaActivationIsReversible(g *game.Game, beforeLibraries [game.NumPlayers]zone.Zone, eventStart int) bool {
	for player, library := range beforeLibraries {
		current := &g.Players[player].Library
		if library.ReorderVersion() != current.ReorderVersion() || !slices.Equal(library.All(), current.All()) {
			return false
		}
	}
	for _, event := range g.Events[eventStart:] {
		if event.Kind == game.EventCardRevealed && event.FromZone == zone.Library ||
			event.Kind == game.EventZoneChanged && (event.FromZone == zone.Library || event.ToZone == zone.Library) {
			return false
		}
	}
	return true
}
