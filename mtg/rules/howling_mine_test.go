package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// eachPlayerDrawStepPermanent builds a Howling Mine-style artifact whose
// triggered ability fires at the beginning of each player's draw step and
// makes that player (the active player) draw an additional card. When
// requireUntapped is true the trigger carries an "if this artifact is
// untapped" intervening condition.
func eachPlayerDrawStepPermanent(g *game.Game, controller game.PlayerID, requireUntapped bool) *game.Permanent {
	ability := game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerAt,
			Pattern: game.TriggerPattern{
				Event: game.EventBeginningOfStep,
				Step:  game.StepDraw,
			},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.Draw{
					Amount: game.Fixed(1),
					Player: game.EventPlayerReference(),
				}},
			},
		}.Ability(),
	}
	if requireUntapped {
		ability.Trigger.InterveningCondition = opt.Val(game.Condition{
			Text:          "if this artifact is untapped",
			Object:        opt.Val(game.SourcePermanentReference()),
			ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Tapped: game.TriFalse}),
		})
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Howling Mine",
		Types:              []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{ability},
	}}
	return addCombatPermanent(g, controller, def)
}

func drawStepExtraDraw(t *testing.T, g *game.Game, engine *Engine, active game.PlayerID) int {
	t.Helper()
	before := g.Players[active].Hand.Size()
	g.Turn.ActivePlayer = active
	emitBeginningOfStepEvent(g, game.StepDraw)
	if engine.putTriggeredAbilitiesOnStack(g) {
		agents := [game.NumPlayers]PlayerAgent{}
		log := TurnLog{}
		engine.resolveTopOfStackWithChoices(g, agents, &log)
	}
	return g.Players[active].Hand.Size() - before
}

func TestEachPlayerDrawStepThatPlayerDraws(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 5)
	fillLibrary(g, game.Player2, 5)

	// Source is controlled by Player1, but the trigger fires on each player's
	// draw step and "that player" resolves to the active player.
	eachPlayerDrawStepPermanent(g, game.Player1, false)

	if got := drawStepExtraDraw(t, g, engine, game.Player1); got != 1 {
		t.Fatalf("Player1 draw step: drew %d extra cards; want 1", got)
	}
	if got := drawStepExtraDraw(t, g, engine, game.Player2); got != 1 {
		t.Fatalf("Player2 draw step: drew %d extra cards; want 1", got)
	}
}

func TestEachPlayerDrawStepUntappedGate(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fillLibrary(g, game.Player1, 5)

	source := eachPlayerDrawStepPermanent(g, game.Player1, true)

	if got := drawStepExtraDraw(t, g, engine, game.Player1); got != 1 {
		t.Fatalf("untapped source: drew %d extra cards; want 1", got)
	}

	source.Tapped = true
	if got := drawStepExtraDraw(t, g, engine, game.Player1); got != 0 {
		t.Fatalf("tapped source: drew %d extra cards; want 0", got)
	}
}
