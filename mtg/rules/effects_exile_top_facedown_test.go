package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// TestExileTopOfLibraryFaceDownPlacesNamedCounter proves the FaceDown field
// exiles each top-of-library card face down while the Counter field stamps the
// card-defined named marker counter, the two riders Flamewar, Streetwise
// Operative applies together ("exile ... face down. Put an intel counter on each
// of them.").
func TestExileTopOfLibraryFaceDownPlacesNamedCounter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ExileTopOfLibrary{
		Amount:   game.Fixed(2),
		Player:   game.ControllerReference(),
		Counter:  opt.Val(counter.Intel),
		FaceDown: true,
	}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	top := []id.ID{
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}}),
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}}),
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	exile := g.Players[game.Player1].Exile
	for _, cardID := range top {
		if !exile.Contains(cardID) {
			t.Fatalf("card %v was not exiled", cardID)
		}
		if !exile.IsFaceDown(cardID) {
			t.Fatalf("exiled card %v is not face down", cardID)
		}
		if !g.HasExileCounter(cardID, counter.Intel) {
			t.Fatalf("exiled card %v is missing its intel counter", cardID)
		}
	}
	if got := g.Players[game.Player1].Exile.Size(); got != 2 {
		t.Fatalf("exile size = %d, want 2", got)
	}
}

// TestExileTopOfLibraryCombatDamageScaledFaceDownCounter proves a combat-damage
// trigger's "that many" dynamic amount feeds the face-down top-of-library exile:
// dealing N combat damage exiles exactly N cards face down, each with the named
// counter (Flamewar, Streetwise Operative: "Whenever Flamewar deals combat
// damage to a player, exile that many cards from the top of your library face
// down. Put an intel counter on each of them.").
func TestExileTopOfLibraryCombatDamageScaledFaceDownCounter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Source:              game.TriggerSourceSelf,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}, []game.Instruction{{Primitive: game.ExileTopOfLibrary{
		Amount:   game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
		Player:   game.ControllerReference(),
		Counter:  opt.Val(counter.Intel),
		FaceDown: true,
	}}}, nil)
	for range 5 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 3, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	exile := g.Players[game.Player1].Exile
	if got := exile.Size(); got != 3 {
		t.Fatalf("exile size = %d, want 3 (equal to combat damage dealt)", got)
	}
	for _, cardID := range exile.All() {
		if !exile.IsFaceDown(cardID) {
			t.Fatalf("exiled card %v is not face down", cardID)
		}
		if !g.HasExileCounter(cardID, counter.Intel) {
			t.Fatalf("exiled card %v is missing its intel counter", cardID)
		}
	}
	if got := g.Players[game.Player1].Library.Size(); got != 2 {
		t.Fatalf("library size = %d, want 2 (5 - 3 exiled)", got)
	}
}

// TestFlamewarExileThenReturnEndToEnd exercises both Flamewar mechanics in
// sequence: the back face's combat-damage trigger exiles cards from the top of
// the owner's library face down with intel counters (and converts), then the
// front face's activated ability returns every intel-countered card the owner
// owns from exile to hand. The returned cards leave exile face up with their
// exile counters cleared.
func TestFlamewarExileThenReturnEndToEnd(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Source:              game.TriggerSourceSelf,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}, []game.Instruction{
		{Primitive: game.ExileTopOfLibrary{
			Amount:   game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
			Player:   game.ControllerReference(),
			Counter:  opt.Val(counter.Intel),
			FaceDown: true,
		}},
		{Primitive: game.Transform{Object: game.SourcePermanentReference()}},
	}, nil)
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	exiled := g.Players[game.Player1].Exile.All()
	if len(exiled) != 2 {
		t.Fatalf("exiled %d cards, want 2", len(exiled))
	}

	addEffectSpellToStack(g, game.Player1, game.ReturnExiledCardsWithCounter{
		Player:  game.ControllerReference(),
		Counter: counter.Intel,
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, cardID := range exiled {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("exiled card %v was not returned to hand", cardID)
		}
		if g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("returned card %v remained in exile", cardID)
		}
		if g.HasExileCounter(cardID, counter.Intel) {
			t.Fatalf("returned card %v kept its intel counter after leaving exile", cardID)
		}
	}
	if got := g.Players[game.Player1].Exile.Size(); got != 0 {
		t.Fatalf("exile size = %d, want 0 after return", got)
	}
}
