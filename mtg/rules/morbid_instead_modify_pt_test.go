package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// morbidInsteadModifyPTInstructions builds Tragic Slip's fused spell sequence:
// a base -1/-1 modification gated on no creature having died this turn and a
// larger -13/-13 modification gated on a creature having died this turn (the
// Morbid "instead" replacement), both applied to the spell's single target.
func morbidInsteadModifyPTInstructions() []game.Instruction {
	creatureDied := game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	}
	noCreatureDied := creatureDied
	noCreatureDied.Negate = true
	modify := func(delta int, condition game.Condition) game.Instruction {
		return game.Instruction{
			Primitive: game.ModifyPT{
				Object:         game.TargetPermanentReference(0),
				PowerDelta:     game.Fixed(delta),
				ToughnessDelta: game.Fixed(delta),
				Duration:       game.DurationUntilEndOfTurn,
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
		}
	}
	return []game.Instruction{
		modify(-1, noCreatureDied),
		modify(-13, creatureDied),
	}
}

// TestMorbidInsteadModifyPTAppliesBaseWhenNoCreatureDied proves Tragic Slip's
// base -1/-1 modification resolves (and its Morbid -13/-13 alternative does not)
// when no creature has died this turn.
func TestMorbidInsteadModifyPTAppliesBaseWhenNoCreatureDied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)

	addInstructionSpellToStackForController(g, game.Player1,
		morbidInsteadModifyPTInstructions(),
		[]game.Target{game.PermanentTarget(creature.ObjectID)},
	)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power = %d, want 4 (5 base - 1)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 4 {
		t.Fatalf("effective toughness = %d (ok=%v), want 4 (5 base - 1)", got, ok)
	}
}

// TestMorbidInsteadModifyPTAppliesMorbidWhenCreatureDied proves Tragic Slip's
// Morbid -13/-13 modification replaces the base -1/-1 when a creature has died
// this turn.
func TestMorbidInsteadModifyPTAppliesMorbidWhenCreatureDied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	emitCreatureDiedEvent(g)

	addInstructionSpellToStackForController(g, game.Player1,
		morbidInsteadModifyPTInstructions(),
		[]game.Target{game.PermanentTarget(creature.ObjectID)},
	)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, creature); got != 0 {
		t.Fatalf("effective power = %d, want 0 (5 base - 13, clamped at 0)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != -8 {
		t.Fatalf("effective toughness = %d (ok=%v), want -8 (5 base - 13)", got, ok)
	}
}
