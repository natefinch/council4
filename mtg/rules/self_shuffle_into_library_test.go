package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestShufflePermanentIntoLibraryMovesDeadCardFromGraveyard covers the dies /
// put-into-graveyard self-recursion "Shuffle it into its owner's library."
// (Alabaster Dragon, Serra Avatar): the permanent has already become a card in
// the graveyard when the ability resolves, so ShufflePermanentIntoLibrary must
// fall back from the (gone) permanent to that card's last-known snapshot and move
// the card from the graveyard into its owner's library.
func TestShufflePermanentIntoLibraryMovesDeadCardFromGraveyard(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, vanillaCreature("Alabaster Dragon", 4, 4))
	cardID := permanent.CardInstanceID
	objectID := permanent.ObjectID

	// The creature has died: snapshot its last-known information, then move its
	// card to the graveyard and remove the permanent from the battlefield.
	snapshot := snapshotPermanent(g, permanent, zone.Graveyard)
	rememberLastKnown(g, &snapshot)
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("failed to move the dead creature's card to the graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("card not in graveyard after death")
	}

	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceID:        objectID,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: objectID},
	}
	resolveInstruction(engine, g, obj, game.ShufflePermanentIntoLibrary{
		Object: game.EventPermanentReference(),
	}, nil)

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("card still in graveyard after shuffle into library")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("card was not shuffled into its owner's library")
	}
}

// TestShufflePermanentIntoLibraryLeavesCardThatLeftGraveyard covers the CR 400.7
// response window: if the dead card leaves the graveyard before the shuffle
// resolves (an opponent exiles it, the owner returns it to hand), it is a new
// object the ability no longer tracks, so the shuffle must not drag it out of its
// new zone into the library.
func TestShufflePermanentIntoLibraryLeavesCardThatLeftGraveyard(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, vanillaCreature("Serra Avatar", 4, 4))
	cardID := permanent.CardInstanceID
	objectID := permanent.ObjectID

	snapshot := snapshotPermanent(g, permanent, zone.Graveyard)
	rememberLastKnown(g, &snapshot)
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("failed to move the dead creature's card to the graveyard")
	}
	// In response, the card is exiled from the graveyard.
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to exile the card from the graveyard")
	}

	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceID:        objectID,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: objectID},
	}
	resolveInstruction(engine, g, obj, game.ShufflePermanentIntoLibrary{
		Object: game.EventPermanentReference(),
	}, nil)

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("card was pulled out of exile instead of staying where it was")
	}
	if g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("card that left the graveyard was wrongly shuffled into the library")
	}
}

// moves a live battlefield permanent into its owner's library, unchanged by the
// dead-card fallback.
func TestShufflePermanentIntoLibraryMovesLivePermanent(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, vanillaCreature("Worldspine Wurm", 15, 15))
	cardID := permanent.CardInstanceID

	obj := &game.StackObject{Controller: game.Player1, SourceID: permanent.ObjectID}
	resolveInstruction(engine, g, obj, game.ShufflePermanentIntoLibrary{
		Object: game.SourcePermanentReference(),
	}, nil)

	if permanentByCardID(g, cardID) != nil {
		t.Fatal("permanent still on battlefield after shuffle into library")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("permanent was not shuffled into its owner's library")
	}
}
