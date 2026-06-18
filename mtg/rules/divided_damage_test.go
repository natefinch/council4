package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// dividedDamage builds a Divided damage instruction whose recipient targets the
// spec at index 0.
func dividedDamage(total int) game.Damage {
	return game.Damage{
		Amount:    game.Fixed(total),
		Recipient: game.AnyTargetDamageRecipient(0),
		Divided:   true,
	}
}

// TestDividedDamageDistributesControllerAllocation proves the controller's
// resolution-time allocation is honored: each chosen target receives the amount
// the player assigned and the allocations sum to the fixed total.
func TestDividedDamageDistributesControllerAllocation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, dividedDamage(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	// Allocate 1 to the first target, 2 to the second (multiset of option indices).
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1, 1}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first target left the battlefield")
	}
	gotSecond, ok := permanentByObjectID(g, second.ObjectID)
	if !ok {
		t.Fatal("second target left the battlefield")
	}
	if gotFirst.MarkedDamage != 1 {
		t.Fatalf("first target marked damage = %d, want 1", gotFirst.MarkedDamage)
	}
	if gotSecond.MarkedDamage != 2 {
		t.Fatalf("second target marked damage = %d, want 2", gotSecond.MarkedDamage)
	}
	if total := gotFirst.MarkedDamage + gotSecond.MarkedDamage; total != 3 {
		t.Fatalf("total damage dealt = %d, want 3 (conserved)", total)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceDamageAllocation {
		t.Fatalf("choices = %+v, want one ChoiceDamageAllocation", log.Choices)
	}
}

// TestDividedDamageDefaultAllocationConservesTotal proves the nil-agent default
// still distributes the full total (one to each target, remainder to the last).
func TestDividedDamageDefaultAllocationConservesTotal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, dividedDamage(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first target left the battlefield")
	}
	gotSecond, ok := permanentByObjectID(g, second.ObjectID)
	if !ok {
		t.Fatal("second target left the battlefield")
	}
	if gotFirst.MarkedDamage < 1 || gotSecond.MarkedDamage < 1 {
		t.Fatalf("default allocation left a target with no damage: first=%d second=%d",
			gotFirst.MarkedDamage, gotSecond.MarkedDamage)
	}
	if total := gotFirst.MarkedDamage + gotSecond.MarkedDamage; total != 3 {
		t.Fatalf("total damage dealt = %d, want 3 (conserved)", total)
	}
}

// TestDividedDamageSpecValidates proves the typed game layer accepts a Divided
// damage instruction whose recipient references an in-range multi-target spec.
func TestDividedDamageSpecValidates(t *testing.T) {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Test Bolt",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 2,
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
				}},
				Sequence: []game.Instruction{{Primitive: dividedDamage(2)}},
			}.Ability()),
		},
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef rejected a valid divided-damage spec: %v", issues)
	}
}
