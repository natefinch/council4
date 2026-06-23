package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestBecomeMonarchEffectSetsSingleMonarch covers the BecomeMonarch primitive
// (CR 720): resolving "you become the monarch" makes the resolving controller
// the monarch and clears any prior monarch so at most one player holds it.
func TestBecomeMonarchEffectSetsSingleMonarch(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player2].IsMonarch = true

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Thorn",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.BecomeMonarch{Player: game.ControllerReference()}, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("controller did not become the monarch")
	}
	if g.Players[game.Player2].IsMonarch {
		t.Fatal("prior monarch was not cleared")
	}
}

// TestBecomeMonarchSkipsEliminatedPlayer confirms an eliminated player cannot
// take the crown and the prior monarch is left unchanged.
func TestBecomeMonarchSkipsEliminatedPlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].IsMonarch = true
	g.Players[game.Player1].Eliminated = true

	if setMonarch(g, game.Player1) {
		t.Fatal("setMonarch unexpectedly crowned an eliminated player")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("eliminated player became the monarch")
	}
	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("prior monarch lost the crown to a failed steal")
	}
}

// TestMonarchDrawsAtEndStep covers CR 720.5: the monarch draws a card at the
// beginning of their end step. A non-monarch active player draws nothing.
func TestMonarchDrawsAtEndStep(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 5)
	g.Turn.ActivePlayer = game.Player1
	g.Players[game.Player1].IsMonarch = true

	before := g.Players[game.Player1].Hand.Size()
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if got := g.Players[game.Player1].Hand.Size() - before; got != 1 {
		t.Fatalf("monarch drew %d cards at end step, want 1", got)
	}

	g2 := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fillLibrary(g2, game.Player1, 5)
	g2.Turn.ActivePlayer = game.Player1
	g2.Players[game.Player2].IsMonarch = true
	beforeNon := g2.Players[game.Player1].Hand.Size()
	engine.runEndingPhase(g2, [game.NumPlayers]PlayerAgent{})
	if got := g2.Players[game.Player1].Hand.Size() - beforeNon; got != 0 {
		t.Fatalf("non-monarch active player drew %d cards at end step, want 0", got)
	}
}

// TestCombatDamageStealsMonarch covers CR 720.6: when a creature deals combat
// damage to the monarch, that creature's controller becomes the monarch.
func TestCombatDamageStealsMonarch(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].IsMonarch = true
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("attacking creature's controller did not become the monarch")
	}
	if g.Players[game.Player2].IsMonarch {
		t.Fatal("prior monarch kept the crown after taking combat damage")
	}
}

// TestCombatDamageToNonMonarchKeepsMonarch confirms combat damage to a player
// who is not the monarch never moves the crown.
func TestCombatDamageToNonMonarchKeepsMonarch(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player3].IsMonarch = true
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})

	if g.Players[game.Player1].IsMonarch {
		t.Fatal("crown moved when combat damage was dealt to a non-monarch")
	}
	if !g.Players[game.Player3].IsMonarch {
		t.Fatal("monarch lost the crown to unrelated combat damage")
	}
}
