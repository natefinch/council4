package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestNecropotenceSkipsControllerDrawStep proves the registered card's first
// Oracle line, "Skip your draw step.", is live on the real definition: its
// controller performs no turn-based draw during the beginning phase, while a
// player who does not control it draws as usual. Driving the actual beginning
// phase rather than the draw handler confirms the static rule effect suppresses
// the turn-based draw the phase would otherwise perform.
func TestNecropotenceSkipsControllerDrawStep(t *testing.T) {
	def := necropotenceCardDef(t)
	static := def.StaticAbilities[0]
	if len(static.RuleEffects) != 1 || static.RuleEffects[0].Kind != game.RuleEffectSkipDrawStep {
		t.Fatalf("static ability rule effects = %+v, want a single skip-draw-step effect", static.RuleEffects)
	}
	if static.RuleEffects[0].AffectedPlayer != game.PlayerYou {
		t.Fatalf("skip-draw affects %v, want the controller (PlayerYou)", static.RuleEffects[0].AffectedPlayer)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addNecropotence(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Controller Draw"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("controller hand size = %d, want 0 (draw step skipped)", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("controller library size = %d, want 1 (turn-based draw not performed)", got)
	}
	if g.Turn.Step == game.StepDraw {
		t.Fatal("turn advanced into the controller's draw step, want it skipped")
	}

	// A player who does not control Necropotence still draws for the turn.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent Draw"}})
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("opponent hand size = %d, want 1 (opponent's draw step is unaffected)", got)
	}
}
