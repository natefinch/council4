package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// multiOriginExileUnionPattern models Laelia, the Blade Reforged's "Whenever one
// or more cards are put into exile from your library and/or your graveyard, ..."
// trigger: a move into exile whose origin is any of the controller's library or
// graveyard, coalesced once per simultaneous batch.
func multiOriginExileUnionPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:       game.EventZoneChanged,
		Player:      game.TriggerPlayerYou,
		FromZones:   []zone.Type{zone.Library, zone.Graveyard},
		MatchToZone: true,
		ToZone:      zone.Exile,
		OneOrMore:   true,
		SubjectSelection: game.Selection{
			NonToken: true,
		},
	}
}

// TestMultiOriginExileUnionFiresForLibraryAndGraveyard verifies the origin union
// fires for a card exiled from the controller's library and for a card exiled
// from the controller's graveyard.
func TestMultiOriginExileUnionFiresForLibraryAndGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, multiOriginExileUnionPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	fromLibrary := addCardToLibrary(g, game.Player1, greenCreature())
	if !moveCardBetweenZones(g, game.Player1, fromLibrary, zone.Library, zone.Exile) {
		t.Fatal("moveCardBetweenZones failed for library card")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger did not fire for a library card")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
	g.Stack.Pop()

	fromGraveyard := addCardToGraveyard(g, game.Player1, greenCreature())
	if !moveCardBetweenZones(g, game.Player1, fromGraveyard, zone.Graveyard, zone.Exile) {
		t.Fatal("moveCardBetweenZones failed for graveyard card")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger did not fire for a graveyard card")
	}
	obj, ok = g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}

// TestMultiOriginExileUnionRejectsOtherOriginsAndPlayers verifies the origin
// union does not fire for a card exiled from a zone outside the union (hand,
// battlefield) nor for a card exiled from an opponent's library.
func TestMultiOriginExileUnionRejectsOtherOriginsAndPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, multiOriginExileUnionPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// A card exiled from your hand must NOT fire (hand is outside the union).
	fromHand := addCardToHand(g, game.Player1, greenCreature())
	if !moveCardBetweenZones(g, game.Player1, fromHand, zone.Hand, zone.Exile) {
		t.Fatal("moveCardBetweenZones failed for hand card")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger fired for a hand card")
	}

	// A permanent exiled from the battlefield must NOT fire.
	permanent := addCombatCreaturePermanent(g, game.Player1)
	if !movePermanentToZone(g, permanent, zone.Exile) {
		t.Fatal("movePermanentToZone failed for battlefield permanent")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger fired for a battlefield permanent")
	}

	// A card exiled from an opponent's library must NOT fire (Player: You).
	opponentLibrary := addCardToLibrary(g, game.Player2, greenCreature())
	if !moveCardBetweenZones(g, game.Player2, opponentLibrary, zone.Library, zone.Exile) {
		t.Fatal("moveCardBetweenZones failed for opponent library card")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger fired for an opponent's library card")
	}
}

// TestMultiOriginExileUnionCoalescesSimultaneousBatch verifies a simultaneous
// batch of qualifying exiles — one from the library and one from the graveyard —
// coalesces into exactly one trigger via one-or-more batching.
func TestMultiOriginExileUnionCoalescesSimultaneousBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, multiOriginExileUnionPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	fromLibrary := addCardToLibrary(g, game.Player1, greenCreature())
	fromGraveyard := addCardToGraveyard(g, game.Player1, greenCreature())
	simultaneousID := g.IDGen.Next()
	if !moveCardBetweenZonesInBatch(g, game.Player1, fromLibrary, zone.Library, zone.Exile, false, simultaneousID) {
		t.Fatal("moveCardBetweenZonesInBatch failed for library card")
	}
	if !moveCardBetweenZonesInBatch(g, game.Player1, fromGraveyard, zone.Graveyard, zone.Exile, false, simultaneousID) {
		t.Fatal("moveCardBetweenZonesInBatch failed for graveyard card")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("multi-origin exile trigger did not fire for a simultaneous batch")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 coalesced trigger for a two-card batch", got)
	}
}
