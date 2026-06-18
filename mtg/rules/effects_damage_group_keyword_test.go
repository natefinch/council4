package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestExcludedKeywordTargetSpecRejectsKeywordBearer exercises the runtime target
// legality of the predicate shape cardgen produces for "target creature without
// flying": a creature lacking the keyword is a legal target while a flyer is
// not.
func TestExcludedKeywordTargetSpecRejectsKeywordBearer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flyer := addCombatCreaturePermanentWithPower(g, game.Player2, 5, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Constraint: "target creature without flying",
		Predicate: game.TargetPredicate{
			PermanentTypes:  []types.Card{types.Creature},
			ExcludedKeyword: game.Flying,
		},
	}

	if !permanentTargetMatchesSpec(g, game.Player1, 0, &spec, grounded.ObjectID) {
		t.Fatal("grounded creature is not a legal target, want legal")
	}
	if permanentTargetMatchesSpec(g, game.Player1, 0, &spec, flyer.ObjectID) {
		t.Fatal("flyer is a legal target for a without-flying spec, want illegal")
	}
}

// TestGroupKeywordDamageHitsOnlyMatchingKeyword verifies that group damage with
// a keyword selector predicate ("each creature with flying") marks damage on
// permanents that have the keyword and leaves other creatures untouched.
func TestGroupKeywordDamageHitsOnlyMatchingKeyword(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flyer := addCombatCreaturePermanentWithPower(g, game.Player2, 5, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Fixed(2),
			Recipient: game.GroupDamageRecipient(
				game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Keyword:       game.Flying,
				}),
			),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	flyerAfter, ok := permanentByObjectID(g, flyer.ObjectID)
	if !ok {
		t.Fatal("flyer not found after resolution")
	}
	if flyerAfter.MarkedDamage != 2 {
		t.Fatalf("flyer marked damage = %d, want 2", flyerAfter.MarkedDamage)
	}
	groundedAfter, ok := permanentByObjectID(g, grounded.ObjectID)
	if !ok {
		t.Fatal("grounded creature not found after resolution")
	}
	if groundedAfter.MarkedDamage != 0 {
		t.Fatalf("grounded creature marked damage = %d, want 0", groundedAfter.MarkedDamage)
	}
}

// TestGroupExcludedKeywordDamageSkipsMatchingKeyword verifies that group damage
// with an excluded-keyword selector ("each creature without flying") marks
// damage on permanents that lack the keyword and leaves keyword-bearers
// untouched.
func TestGroupExcludedKeywordDamageSkipsMatchingKeyword(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flyer := addCombatCreaturePermanentWithPower(g, game.Player2, 5, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Fixed(2),
			Recipient: game.GroupDamageRecipient(
				game.BattlefieldGroup(game.Selection{
					RequiredTypes:   []types.Card{types.Creature},
					ExcludedKeyword: game.Flying,
				}),
			),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	flyerAfter, ok := permanentByObjectID(g, flyer.ObjectID)
	if !ok {
		t.Fatal("flyer not found after resolution")
	}
	if flyerAfter.MarkedDamage != 0 {
		t.Fatalf("flyer marked damage = %d, want 0", flyerAfter.MarkedDamage)
	}
	groundedAfter, ok := permanentByObjectID(g, grounded.ObjectID)
	if !ok {
		t.Fatal("grounded creature not found after resolution")
	}
	if groundedAfter.MarkedDamage != 2 {
		t.Fatalf("grounded creature marked damage = %d, want 2", groundedAfter.MarkedDamage)
	}
}
