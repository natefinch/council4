package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestSkipDrawStepStaticSkipsTurnBasedDraw confirms that a battlefield permanent
// carrying the "Skip your draw step." static rule (Necropotence, Yawgmoth's
// Bargain) makes its controller's draw step not happen: the turn-based draw is
// not performed during the active player's beginning phase.
func TestSkipDrawStepStaticSkipsTurnBasedDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Skip Draw Enchantment",
		StaticAbilities: []game.StaticAbility{game.SkipDrawStepStaticBody},
	}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want 0 (draw step skipped)", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1 (turn-based draw not performed)", got)
	}
	if g.Turn.Step == game.StepDraw {
		t.Fatal("turn advanced into the draw step, want it skipped")
	}
}

// TestSkipDrawStepStaticConsumesScheduledSkip confirms that a scheduled one-shot
// draw-step skip is consumed during the active player's beginning phase even
// when a static "Skip your draw step." effect also skips the step. The queued
// skip must not survive to skip a later, unintended draw step.
func TestSkipDrawStepStaticConsumesScheduledSkip(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Skip Draw Enchantment",
		StaticAbilities: []game.StaticAbility{game.SkipDrawStepStaticBody},
	}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	scheduleSkipStep(g, game.Player1, game.StepDraw)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if consumeSkipStep(g, game.Player1, game.StepDraw) {
		t.Fatal("scheduled draw-step skip survived; it should have been consumed by the already-skipped step")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want 0 (draw step skipped)", got)
	}
}

// TestSkipDrawStepStaticSkipsTurnBasedDraw: with no skip-draw-step static in
// play, the active player performs the turn-based draw as usual.
func TestDrawStepHappensWithoutSkipStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 (turn-based draw performed)", got)
	}
}
