package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPastStackConditionUsesLogicalIndexAfterTargetCompaction(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Gated Numeric", Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Gate: game.TargetGateSpellKicked},
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject},
		}, Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(1)}}}}.Ability())}}
	cardID := addCardToHand(g, game.Player1, def)
	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1, SourceID: cardID,
		Targets: []game.Target{game.StackObjectTarget(target.ID)}}
	if remapTargetSlot(g, obj, 1) != 0 {
		t.Fatal("fixture did not compact logical target")
	}
	if !counterTargetStackObject(g, obj, 1, false, game.CounteredSpellGraveyard) {
		t.Fatal("counter failed")
	}
	condition := stackSelectionCondition(1, game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 3})})
	condition.Condition.Val.UseCounteredSpellManaValue = true
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition.Condition) {
		t.Fatal("logical counter information was looked up by physical slot")
	}
	if _, wrong := obj.TargetManaValueLKI[0]; wrong {
		t.Fatal("capture used physical slot")
	}
}

func TestMalformedCounteredSpellInformationFailsClosed(t *testing.T) {
	t.Parallel()
	for _, negate := range []bool{false, true} {
		condition := opt.Val(game.Condition{UseCounteredSpellManaValue: true, Negate: negate})
		if conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{}, condition) {
			t.Fatal("flag-only condition evaluated successfully")
		}
	}
}
