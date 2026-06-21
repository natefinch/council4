package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestMetalcraftBackReferenceExileGatesOnControlCount proves the Tap-then-exile
// sequence the cardgen backend emits for Dispatch ("Tap target creature.
// Metalcraft — If you control three or more artifacts, exile that creature.")
// always taps the targeted creature but only exiles it — the conditionally
// gated back-reference instruction addressing the same target index — when the
// controller satisfies the Metalcraft control-count condition.
func TestMetalcraftBackReferenceExileGatesOnControlCount(t *testing.T) {
	dispatchSequence := func() []game.Instruction {
		return []game.Instruction{
			{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
			{
				Primitive: game.Exile{Object: game.TargetPermanentReference(0)},
				Condition: opt.Val(game.EffectCondition{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MinCount:  3,
						}),
					}),
				}),
			},
		}
	}

	artifact := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:  "Relic",
			Types: []types.Card{types.Artifact},
		}}
	}

	t.Run("metalcraft active exiles the tapped creature", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		for range 3 {
			addCombatPermanent(g, game.Player1, artifact())
		}
		creature := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, dispatchSequence(), []game.Target{
			game.PermanentTarget(creature.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
			t.Fatal("creature remained on battlefield with three artifacts")
		}
		if !g.Players[game.Player2].Exile.Contains(creature.CardInstanceID) {
			t.Fatal("creature was not exiled to its owner's exile zone")
		}
	})

	t.Run("metalcraft inactive only taps the creature", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		for range 2 {
			addCombatPermanent(g, game.Player1, artifact())
		}
		creature := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, dispatchSequence(), []game.Target{
			game.PermanentTarget(creature.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		survivor, ok := permanentByObjectID(g, creature.ObjectID)
		if !ok {
			t.Fatal("creature was exiled despite only two artifacts")
		}
		if !survivor.Tapped {
			t.Fatal("creature was not tapped by the unconditional tap effect")
		}
		if g.Players[game.Player2].Exile.Contains(creature.CardInstanceID) {
			t.Fatal("creature was exiled despite failing the Metalcraft condition")
		}
	})
}
