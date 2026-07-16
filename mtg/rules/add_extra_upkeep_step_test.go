package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// paradoxHazeAuraCard builds a minimal Paradox Haze: an "Enchant player" Aura
// whose triggered ability fires at the beginning of the enchanted player's first
// upkeep each turn and inserts one additional upkeep step. It exercises the
// additional-upkeep-step engine without depending on the generated card body.
func paradoxHazeAuraCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Paradox Haze",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			game.EnchantStaticAbility(&game.TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "player",
				Allow:      game.TargetAllowPlayer,
			}),
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:                             game.EventBeginningOfStep,
					Step:                              game.StepUpkeep,
					StepPlayerIsSourceEnchantedPlayer: true,
					FirstUpkeepStepEachTurn:           true,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddExtraUpkeepStep{},
			}}}.Ability(),
		}},
	}}
}

// addParadoxHazeAura places a resolved Paradox Haze Aura permanent on the
// battlefield under controller, already attached to the enchanted player.
func addParadoxHazeAura(g *game.Game, controller, enchanted game.PlayerID) *game.Permanent {
	aura := addCombatPermanent(g, controller, paradoxHazeAuraCard())
	aura.AttachedToPlayer = opt.Val(enchanted)
	return aura
}

// countSteps returns how many logged steps match the given step.
func countSteps(log *TurnLog, step game.Step) int {
	count := 0
	for _, entry := range log.Steps {
		if entry.Step == step {
			count++
		}
	}
	return count
}

// TestAddExtraUpkeepStepPrimitiveIncrements proves resolving an AddExtraUpkeepStep
// effect increments TurnState.ExtraUpkeepSteps, and that multiple resolutions
// stack (multiple Paradox Hazes each add a step).
func TestAddExtraUpkeepStepPrimitiveIncrements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addEffectSpellToStack(g, game.Player1, game.AddExtraUpkeepStep{}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Turn.ExtraUpkeepSteps != 1 {
		t.Fatalf("ExtraUpkeepSteps = %d, want 1", g.Turn.ExtraUpkeepSteps)
	}

	addEffectSpellToStack(g, game.Player1, game.AddExtraUpkeepStep{}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Turn.ExtraUpkeepSteps != 2 {
		t.Fatalf("ExtraUpkeepSteps = %d, want 2 (multiple Hazes stack)", g.Turn.ExtraUpkeepSteps)
	}
}

// TestParadoxHazeInsertsUpkeepStepBeforeDraw proves a Paradox Haze attached to the
// active player inserts exactly one additional upkeep step, running it after the
// normal upkeep and before the draw step, and that the additional upkeep does not
// retrigger the "first upkeep each turn" ability (so the loop terminates).
func TestParadoxHazeInsertsUpkeepStepBeforeDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addParadoxHazeAura(g, game.Player1, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	log := TurnLog{}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &log)

	if got := countSteps(&log, game.StepUpkeep); got != 2 {
		t.Fatalf("upkeep steps = %d, want 2 (normal + one additional)", got)
	}
	if g.Turn.UpkeepStepsThisTurn != 2 {
		t.Fatalf("UpkeepStepsThisTurn = %d, want 2", g.Turn.UpkeepStepsThisTurn)
	}
	if g.Turn.ExtraUpkeepSteps != 0 {
		t.Fatalf("ExtraUpkeepSteps = %d, want 0 (drained)", g.Turn.ExtraUpkeepSteps)
	}
	// The draw step must run after both upkeep steps.
	lastUpkeep, firstDraw := -1, -1
	for i, entry := range log.Steps {
		if entry.Step == game.StepUpkeep {
			lastUpkeep = i
		}
		if entry.Step == game.StepDraw && firstDraw == -1 {
			firstDraw = i
		}
	}
	if firstDraw == -1 {
		t.Fatal("draw step did not run")
	}
	if firstDraw < lastUpkeep {
		t.Fatalf("draw step (index %d) ran before the last upkeep step (index %d)", firstDraw, lastUpkeep)
	}
}

// TestMultipleParadoxHazesStackUpkeepSteps proves two Paradox Hazes on the same
// active player each add an upkeep step, yielding three upkeep steps total.
func TestMultipleParadoxHazesStackUpkeepSteps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addParadoxHazeAura(g, game.Player1, game.Player1)
	addParadoxHazeAura(g, game.Player1, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	log := TurnLog{}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &log)

	if got := countSteps(&log, game.StepUpkeep); got != 3 {
		t.Fatalf("upkeep steps = %d, want 3 (normal + two additional)", got)
	}
	if g.Turn.UpkeepStepsThisTurn != 3 {
		t.Fatalf("UpkeepStepsThisTurn = %d, want 3", g.Turn.UpkeepStepsThisTurn)
	}
}

// TestParadoxHazeOnActivePlayerControlledByOpponent proves the trigger is scoped
// to the enchanted player, not the Aura's controller: a Paradox Haze controlled
// by the opponent but attached to the active player still fires on the active
// (enchanted) player's upkeep.
func TestParadoxHazeOnActivePlayerControlledByOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addParadoxHazeAura(g, game.Player2, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	log := TurnLog{}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &log)

	if got := countSteps(&log, game.StepUpkeep); got != 2 {
		t.Fatalf("upkeep steps = %d, want 2 (enchanted player is the active player)", got)
	}
}

// TestParadoxHazeAttachedToNonActivePlayerDoesNotTrigger proves the trigger does
// not fire on the active player's upkeep when the Aura enchants a different
// player, because the upkeep step's event player is the active player and the
// pattern routes through the source's enchanted player.
func TestParadoxHazeAttachedToNonActivePlayerDoesNotTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addParadoxHazeAura(g, game.Player2, game.Player2)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	log := TurnLog{}
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &log)

	if got := countSteps(&log, game.StepUpkeep); got != 1 {
		t.Fatalf("upkeep steps = %d, want 1 (enchanted player is not the active player)", got)
	}
	if g.Turn.ExtraUpkeepSteps != 0 {
		t.Fatalf("ExtraUpkeepSteps = %d, want 0", g.Turn.ExtraUpkeepSteps)
	}
}
