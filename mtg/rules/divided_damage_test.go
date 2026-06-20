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

// TestDividedDamageDropsShareOfIllegalTarget proves the CR 601.2d rule that a
// target which has become illegal since the division was chosen is dealt no
// damage and its assigned share is lost, never redistributed to the surviving
// targets. The controller assigned 1 to the first target and 2 to the second;
// the second is removed before resolution, so only 1 total is dealt (to the
// first), NOT the full 3 concentrated onto the survivor.
func TestDividedDamageDropsShareOfIllegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, dividedDamage(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	// Division over both originally chosen targets: 1 to the first, 2 to the
	// second (multiset of option indices over the original target order).
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1, 1}}},
	}
	// The second target becomes illegal in response (e.g. sacrificed) before the
	// divided-damage spell resolves.
	if _, ok := removePermanentFromBattlefield(g, second.ObjectID); !ok {
		t.Fatal("failed to remove the second target before resolution")
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first (still-legal) target left the battlefield")
	}
	if gotFirst.MarkedDamage != 1 {
		t.Fatalf("surviving target marked damage = %d, want 1 (its assigned share, not redistributed)", gotFirst.MarkedDamage)
	}
	if _, stillThere := permanentByObjectID(g, second.ObjectID); stillThere {
		t.Fatal("removed target unexpectedly still on the battlefield")
	}
	// Total dealt is strictly less than the fixed total: the 2 assigned to the
	// illegal target is lost rather than redistributed.
	if gotFirst.MarkedDamage >= 3 {
		t.Fatalf("dropped share was redistributed: surviving target took %d, want only its assigned 1", gotFirst.MarkedDamage)
	}
}

func TestDividedDamageDropsShareOfPhasedOutTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, dividedDamage(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1, 1}}},
	}
	second.PhasedOut = true

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if first.MarkedDamage != 1 {
		t.Fatalf("active target marked damage = %d, want 1", first.MarkedDamage)
	}
	if second.MarkedDamage != 0 {
		t.Fatalf("phased-out target marked damage = %d, want 0", second.MarkedDamage)
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
