package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestTwoHideawaySourcesPlayOnlyOwnCard proves two Hideaway lands controlled by
// the same player keep independent source-scoped links: playing from one land
// casts only that land's hidden card and leaves the other land's card exiled and
// still linked.
func TestTwoHideawaySourcesPlayOnlyOwnCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landA := addHideawayLandPermanent(g, game.Player1)
	landB := addHideawayLandPermanent(g, game.Player1)

	// Library top-to-bottom: cardA then cardB. Each land exiles the current top.
	cardB := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Hidden B"))
	cardA := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Hidden A"))

	objA := hideawaySourceObject(landA)
	objB := hideawaySourceObject(landB)
	engine.resolveInstructionWithChoices(g, objA, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, objB, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	keyA := linkedObjectSourceKey(g, objA, hideawayLinkID)
	keyB := linkedObjectSourceKey(g, objB, hideawayLinkID)
	if refs := linkedObjects(g, keyA); len(refs) != 1 || refs[0].CardID != cardA {
		t.Fatalf("land A link = %+v, want single ref to %v", refs, cardA)
	}
	if refs := linkedObjects(g, keyB); len(refs) != 1 || refs[0].CardID != cardB {
		t.Fatalf("land B link = %+v, want single ref to %v", refs, cardB)
	}

	// Playing from land A casts only card A.
	engine.resolveInstructionWithChoices(g, objA, &game.Instruction{Primitive: game.PlayHideawayCard{}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardA {
		t.Fatalf("stack top = %+v, want land A's hidden card %v", top, cardA)
	}
	if g.Players[game.Player1].Exile.Contains(cardA) {
		t.Fatal("land A's hidden card should have left exile when played")
	}
	if refs := linkedObjects(g, keyA); len(refs) != 0 {
		t.Fatalf("land A link not cleared after play: %+v", refs)
	}
	// Land B is untouched: its card stays exiled and linked.
	if !g.Players[game.Player1].Exile.Contains(cardB) {
		t.Fatal("land B's hidden card must remain exiled")
	}
	if refs := linkedObjects(g, keyB); len(refs) != 1 || refs[0].CardID != cardB {
		t.Fatalf("land B link = %+v, want it intact after playing from land A", refs)
	}
}

// TestPlayHideawayCardHiddenCardMovedAwayLeavesStackEmpty proves that when the
// linked card has already left exile, the play ability does nothing rather than
// casting a card from the wrong zone.
func TestPlayHideawayCardHiddenCardMovedAwayLeavesStackEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addHideawayLandPermanent(g, game.Player1)
	spell := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Hidden Spell"))
	obj := hideawaySourceObject(land)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	// The hidden card is moved out of exile before the play ability resolves.
	g.Players[game.Player1].Exile.Remove(spell)
	g.Players[game.Player1].Graveyard.Add(spell)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.PlayHideawayCard{}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want nothing cast when the hidden card left exile", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("moved card should remain where it went, untouched by the play ability")
	}
}
