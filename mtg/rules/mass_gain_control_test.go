package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addEnchantmentPermanent places a plain enchantment permanent under controller,
// for exercising battlefield-wide mass gain-control (Aura Thief).
func addEnchantmentPermanent(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Enchantment},
	}})
}

// massControlInstruction builds the lowered mass gain-control instruction: a
// permanent-duration LayerControl continuous effect whose new controller is the
// resolving controller (the Player1 sentinel) carried on the given group.
func massControlInstruction(group game.GroupReference) game.Instruction {
	return game.Instruction{
		Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:         game.LayerControl,
				NewController: opt.Val(game.Player1),
				Group:         group,
			}},
			Duration: game.DurationPermanent,
		},
	}
}

// TestMassControlBattlefieldGroupGainsControlOfEveryMatchingPermanent proves the
// unqualified-selector sibling form (Aura Thief's "gain control of all
// enchantments"): the resolving controller gains control of every enchantment on
// the battlefield regardless of its current controller, while non-enchantments
// are left alone.
func TestMassControlBattlefieldGroupGainsControlOfEveryMatchingPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	mine := addEnchantmentPermanent(g, game.Player1, "Ghostly Prison")
	theirs := addEnchantmentPermanent(g, game.Player2, "Rancor")
	third := addEnchantmentPermanent(g, game.Player3, "Oblivion Ring")
	creature := addCreaturePermanent(g, game.Player2)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		massControlInstruction(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Enchantment},
		})),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, enchantment := range []*game.Permanent{mine, theirs, third} {
		if got := effectiveController(g, enchantment); got != game.Player1 {
			t.Fatalf("enchantment controller = %v, want Player1 (all enchantments)", got)
		}
	}
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("creature controller = %v, want Player2 (not an enchantment)", got)
	}
}

// TestMassControlTargetPlayerGroupGainsOnlyThatPlayersPermanents proves the
// targeted-player sibling form ("gain control of all creatures target opponent
// controls", Ashiok, Sculptor of Fears): only the targeted player's creatures
// are taken; another player's creatures stay put.
func TestMassControlTargetPlayerGroupGainsOnlyThatPlayersPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	targeted := addCreaturePermanent(g, game.Player2)
	otherTargeted := addCreaturePermanent(g, game.Player2)
	bystander := addCreaturePermanent(g, game.Player3)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		massControlInstruction(game.PlayerControlledGroup(
			game.TargetPlayerReference(0),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
		)),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, stolen := range []*game.Permanent{targeted, otherTargeted} {
		if got := effectiveController(g, stolen); got != game.Player1 {
			t.Fatalf("targeted player's creature controller = %v, want Player1", got)
		}
	}
	if got := effectiveController(g, bystander); got != game.Player3 {
		t.Fatalf("bystander creature controller = %v, want Player3 (only the targeted player)", got)
	}
}

// TestMassControlLatestTimestampControlsPermanent proves control effects apply in
// timestamp order (CR 613.7a): when two mass gain-control effects both claim the
// same permanent, the later effect wins. Player1 steals Player2's enchantment,
// then Player3's later mass gain-control takes it from Player1.
func TestMassControlLatestTimestampControlsPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	enchantment := addEnchantmentPermanent(g, game.Player2, "Sylvan Library")

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		massControlInstruction(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Enchantment},
		})),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, enchantment); got != game.Player1 {
		t.Fatalf("controller after first steal = %v, want Player1", got)
	}

	addInstructionSpellToStackForController(g, game.Player3, []game.Instruction{
		massControlInstruction(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Enchantment},
		})),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, enchantment); got != game.Player3 {
		t.Fatalf("controller after later steal = %v, want Player3 (latest timestamp wins)", got)
	}
}
