package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestSameNamePumpModifiesTargetAndSameNamedCreatures proves that the lowered
// Bile Blight / Echoing-pump shape — an until-end-of-turn LayerPowerToughnessModify
// continuous effect over a SameNamePermanentGroup anchored on the chosen target —
// modifies the target together with every other creature sharing its name, while
// sparing a differently-named creature ("Target creature and all other creatures
// with the same name as that creature get -3/-3 until end of turn."). The
// same-named creature under a different controller is affected too, since the
// group binds on name, not controller.
func TestSameNamePumpModifiesTargetAndSameNamedCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	makeCreature := func(name string, controller game.PlayerID, pt int) *game.Permanent {
		v := game.PT{Value: pt}
		return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(v),
			Toughness: opt.Val(v),
		}})
	}

	// Two 5/5 "Grizzly Bears" under different controllers plus a differently
	// named "Runeclaw Bear" that must be spared.
	anchor := makeCreature("Grizzly Bears", game.Player1, 5)
	sameName := makeCreature("Grizzly Bears", game.Player2, 5)
	different := makeCreature("Runeclaw Bear", game.Player1, 5)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerPowerToughnessModify,
				Group: game.SameNamePermanentGroup(
					game.TargetPermanentReference(0),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				PowerDelta:     -3,
				ToughnessDelta: -3,
			}},
			Duration: game.DurationUntilEndOfTurn,
		}},
	}, []game.Target{game.PermanentTarget(anchor.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	assertPT := func(label string, permanent *game.Permanent, wantPower, wantToughness int) {
		t.Helper()
		if got := effectivePower(g, permanent); got != wantPower {
			t.Fatalf("%s effective power = %d, want %d", label, got, wantPower)
		}
		got, ok := effectiveToughness(g, permanent)
		if !ok {
			t.Fatalf("%s effective toughness unavailable", label)
		}
		if got != wantToughness {
			t.Fatalf("%s effective toughness = %d, want %d", label, got, wantToughness)
		}
	}

	// Both same-named creatures shrink 5/5 -> 2/2; the differently named one
	// keeps its printed 5/5.
	assertPT("anchor", anchor, 2, 2)
	assertPT("same-name", sameName, 2, 2)
	assertPT("different", different, 5, 5)
}
