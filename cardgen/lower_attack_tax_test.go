package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLowerAttackTaxUntilYourNextTurn proves that the resolving,
// duration-bounded attack-tax chapter "Until your next turn, creatures can't
// attack you unless their controller pays {2} for each of those creatures."
// (Summon: Yojimbo chapters II/III) lowers to an ApplyRule carrying a
// controller-scoped RuleEffectAttackTax with an until-your-next-turn duration.
func TestLowerAttackTaxUntilYourNextTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Yojimbo Tax",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Draw a card.\n" +
			"II — Draw a card.\n" +
			"III — Until your next turn, creatures can't attack you unless their controller pays {2} for each of those creatures.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[2].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("chapter III sequence len = %d, want 1", len(mode.Sequence))
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("duration = %v, want DurationUntilYourNextTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectAttackTax {
		t.Fatalf("kind = %v, want RuleEffectAttackTax", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.AttackTaxGeneric != 2 {
		t.Fatalf("attack tax generic = %d, want 2", effect.AttackTaxGeneric)
	}
}

// TestLowerOpponentControllingCountTreasure proves that the where-X variable
// Treasure chapter "Create X Treasure tokens, where X is the number of
// opponents who control a creature with power 4 or greater." (Summon: Yojimbo
// chapter IV) lowers to a CreateToken whose count is the per-opponent
// DynamicAmountOpponentControllingCount, scoped by a "you control a creature
// with power 4 or greater" battlefield group evaluated relative to each
// opponent.
func TestLowerOpponentControllingCountTreasure(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Yojimbo Treasure",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Draw a card.\n" +
			"II — Draw a card.\n" +
			"III — Create X Treasure tokens, where X is the number of opponents who control a creature with power 4 or greater.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[2].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("chapter III sequence len = %d, want 1", len(mode.Sequence))
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", mode.Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok || def.Name != string(types.Treasure) {
		t.Fatalf("token def = %#v, want a Treasure token", create.Source)
	}
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %d, want a dynamic per-opponent count", create.Amount.Value())
	}
	want := game.DynamicAmount{
		Kind:       game.DynamicAmountOpponentControllingCount,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
			Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
		}),
	}
	if got := create.Amount.DynamicAmount().Val; got.Kind != want.Kind ||
		got.Multiplier != want.Multiplier {
		t.Fatalf("dynamic amount = %+v, want kind/multiplier %+v", got, want)
	}
	if got := create.Amount.DynamicAmount().Val.Group; !reflect.DeepEqual(got, want.Group) {
		t.Fatalf("dynamic group = %+v, want %+v", got, want.Group)
	}
}
