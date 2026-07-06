package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// monarchEndStepDrawPermanent builds an Archivist of Gondor-style trigger: "At
// the beginning of the monarch's end step, that player draws a card." The
// pattern is scoped to the monarch (TriggerPlayerMonarch, matched against the
// step's active player) and "that player" resolves to that same player via
// EventPlayerReference.
func monarchEndStepDrawPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Archivist",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:  game.EventBeginningOfStep,
					Step:   game.StepEnd,
					Player: game.TriggerPlayerMonarch,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{
					Amount: game.Fixed(1),
					Player: game.EventPlayerReference(),
				},
			}}}.Ability(),
		}},
	}}
	return addCombatPermanent(g, controller, def)
}

func monarchEndStepDraw(t *testing.T, g *game.Game, engine *Engine, active game.PlayerID) int {
	t.Helper()
	before := g.Players[active].Hand.Size()
	g.Turn.ActivePlayer = active
	emitBeginningOfStepEvent(g, game.StepEnd)
	if engine.putTriggeredAbilitiesOnStack(g) {
		agents := [game.NumPlayers]PlayerAgent{}
		log := TurnLog{}
		engine.resolveTopOfStackWithChoices(g, agents, &log)
	}
	return g.Players[active].Hand.Size() - before
}

// TestMonarchEndStepDrawFiresOnMonarchEndStep proves the monarch-scoped
// beginning-of-end-step trigger fires only on the monarch's end step, and that
// the drawing player ("that player") is the monarch whose end step it is —
// regardless of who controls the source.
func TestMonarchEndStepDrawFiresOnMonarchEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 5)
	fillLibrary(g, game.Player2, 5)

	// Player1 controls the source; Player2 is the monarch.
	monarchEndStepDrawPermanent(g, game.Player1)
	g.Players[game.Player2].IsMonarch = true

	// On the monarch's (Player2's) end step the trigger fires and the monarch
	// draws a card.
	if got := monarchEndStepDraw(t, g, engine, game.Player2); got != 1 {
		t.Fatalf("monarch end step: drew %d cards; want 1", got)
	}

	// On a non-monarch's (Player1's) end step the trigger does not fire.
	if got := monarchEndStepDraw(t, g, engine, game.Player1); got != 0 {
		t.Fatalf("non-monarch end step: drew %d cards; want 0", got)
	}
}

// TestMonarchEndStepDrawSkipsWhenNoMonarch proves the trigger stays inert while
// there is no monarch: an end step with no monarch matches no player.
func TestMonarchEndStepDrawSkipsWhenNoMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 5)

	monarchEndStepDrawPermanent(g, game.Player1)

	if got := monarchEndStepDraw(t, g, engine, game.Player1); got != 0 {
		t.Fatalf("end step with no monarch: drew %d cards; want 0", got)
	}
}
