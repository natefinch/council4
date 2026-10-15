package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func departedPlayer(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	return ok && player.Eliminated
}

// Keep the card registry as historical definition/LKI storage, but remove every
// live incarnation without running ordinary zone replacements (CR 800.4a).
func leaveDepartedOwnedObjects(g *game.Game) {
	var permanents []*game.Permanent
	var events []game.Event
	liveCards := make(map[id.ID]bool)
	batch := g.IDGen.Next()
	for _, permanent := range g.Battlefield {
		if !departedPlayer(g, permanent.Owner) {
			continue
		}
		snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
		rememberLastKnown(g, &snapshot)
		permanents = append(permanents, permanent)
		liveCards[permanent.CardInstanceID] = true
		for _, component := range permanent.MergedCards {
			liveCards[component.CardInstanceID] = true
		}
		if permanent.PhasedOut {
			continue
		}
		events = append(events, prepareEvent(g, game.Event{
			Kind: game.EventObjectLeftGame, Controller: snapshot.Controller, Player: permanent.Owner,
			CardID: permanent.CardInstanceID, PermanentID: permanent.ObjectID,
			Face: permanent.Face, FaceDown: permanent.FaceDown,
			TokenDef: permanent.TokenDef, TokenName: permanentTokenName(permanent),
			FromZone: zone.Battlefield, SimultaneousID: batch,
		}))
	}
	for _, permanent := range permanents {
		if permanent.FaceDown {
			emitFaceDownRevealEvent(g, permanent)
		}
		detachPermanent(g, permanent)
		detachAttachmentsFromPermanent(g, permanent)
		removePermanentFromCombat(g, permanent.ObjectID)
		removePermanentFromBattlefield(g, permanent.ObjectID)
		delete(g.CardCounters, permanent.ObjectID)
	}
	for _, obj := range g.Stack.Objects() {
		owner, known := stackObjectOwner(g, obj)
		if known && departedPlayer(g, owner) {
			g.Stack.RemoveByID(obj.ID)
			if obj.Kind == game.StackSpell && !obj.Copy {
				liveCards[obj.SourceID] = true
			}
		}
	}
	for cardID, card := range g.CardInstances {
		if !departedPlayer(g, card.Owner) {
			continue
		}
		removed := false
		for _, player := range g.Players {
			for _, cards := range []*zone.Zone{&player.Library, &player.Hand, &player.Graveyard, &player.Exile, &player.CommandZone} {
				removed = cards.Remove(cardID) || removed
			}
		}
		if removed || liveCards[cardID] {
			card.ZoneVersion++
		}
		delete(g.CardCounters, cardID)
		delete(g.AdventureCards, cardID)
		delete(g.SuspendedCards, cardID)
		delete(g.ReboundCards, cardID)
		delete(g.PlottedCards, cardID)
		delete(g.ForetoldCards, cardID)
		delete(g.ExileCounterExiledBy, cardID)
	}
	g.Emblems = slices.DeleteFunc(g.Emblems, func(emblem game.Emblem) bool {
		return departedPlayer(g, emblem.Owner)
	})
	for _, event := range events {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			event.CardZoneVersion = card.ZoneVersion
		}
		emitPreparedEvent(g, event)
	}
}

func stackObjectOwner(g *game.Game, obj *game.StackObject) (game.PlayerID, bool) {
	if obj.Copy {
		if obj.CopyOwner.Exists {
			return obj.CopyOwner.Val, true
		}
		return obj.Controller, true
	}
	if obj.Kind != game.StackSpell {
		if obj.AbilityOwner.Exists {
			return obj.AbilityOwner.Val, true
		}
		return obj.Controller, true
	}
	card, ok := g.GetCardInstance(obj.SourceID)
	if ok {
		return card.Owner, true
	}
	return 0, false
}

func cleanupDepartedStackObjects(g *game.Game, playerID game.PlayerID) {
	for _, obj := range g.Stack.Objects() {
		if obj.Controller != playerID {
			continue
		}
		g.Stack.RemoveByID(obj.ID)
		if obj.Kind == game.StackSpell && !obj.Copy {
			if card, ok := g.GetCardInstance(obj.SourceID); ok && !departedPlayer(g, card.Owner) {
				if !moveStackCardToZone(g, obj, card, zone.Exile, false) {
					panic("departed controller's card spell could not leave the stack")
				}
			}
		}
	}
}

// A departed default controller is not replaced by the owner. If no surviving
// control effect remains, exile the object immediately (CR 800.4a, 800.4c).
func settleDepartedControl(g *game.Game) bool {
	if !slices.ContainsFunc(g.Players[:], func(player *game.Player) bool { return player.Eliminated }) {
		return false
	}
	changed := false
	limit := len(g.Battlefield) + len(g.ContinuousEffects) + 1
	for range limit {
		changed = removeExpiredConditionalControlDurations(g) || changed
		var moves []preparedPermanentZoneMove
		for _, permanent := range g.Battlefield {
			if departedPlayer(g, permanent.Owner) || !departedPlayer(g, effectiveController(g, permanent)) {
				continue
			}
			if move, ok := preparePermanentZoneMove(g, permanent, zone.Exile); ok {
				moves = append(moves, move)
			}
		}
		moved := false
		batch := id.ID(0)
		if len(moves) > 0 {
			batch = g.IDGen.Next()
		}
		for i := range moves {
			moves[i].event.SimultaneousID = batch
			moved = applyPreparedPermanentZoneMove(g, &moves[i]) || moved
		}
		if !moved {
			return changed
		}
		changed = true
	}
	panic("departed-player control cleanup did not converge")
}
