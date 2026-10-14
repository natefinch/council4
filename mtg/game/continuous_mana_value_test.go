package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestManaValueBaseCharacteristicsValidateOnlyUnambiguousLayer7b(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		effect ContinuousEffect
		valid  bool
	}{
		{name: "power only", effect: ContinuousEffect{Layer: LayerPowerToughnessSet, SetPowerToManaValue: true}, valid: true},
		{name: "toughness only", effect: ContinuousEffect{Layer: LayerPowerToughnessSet, SetToughnessToManaValue: true}, valid: true},
		{name: "both", effect: ContinuousEffect{Layer: LayerPowerToughnessSet, SetPowerToManaValue: true, SetToughnessToManaValue: true}, valid: true},
		{name: "wrong layer", effect: ContinuousEffect{Layer: LayerPowerToughnessModify, SetPowerToManaValue: true}},
		{name: "competing fixed", effect: ContinuousEffect{Layer: LayerPowerToughnessSet, SetPowerToManaValue: true, SetPower: opt.Val(PT{Value: 1})}},
		{name: "competing dynamic", effect: ContinuousEffect{Layer: LayerPowerToughnessSet, SetToughnessToManaValue: true, SetToughnessDynamic: opt.Val(DynamicAmount{Kind: DynamicAmountConstant, Constant: 1})}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := test.effect
			effect.Group = BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Enchantment}})
			def := &CardDef{CardFace: CardFace{Name: "Mana-value static", StaticAbilities: []StaticAbility{{ContinuousEffects: []ContinuousEffect{effect}}}}}
			if issues := ValidateCardDef(def); (len(issues) == 0) != test.valid {
				t.Fatalf("valid=%v issues=%v", test.valid, issues)
			}
		})
	}
}

func TestReplacementProgramFirstDrawExemptionIsApplicabilityNotAction(t *testing.T) {
	t.Parallel()
	effect := ReplacementEffect{
		MatchEvent: EventCardDrawn, Program: ProgramReference{ID: "test/chains", Version: "1"},
		DrawCardExceptFirstInDrawStep: true,
	}
	if err := effect.ValidateProgram(); err != nil {
		t.Fatal(err)
	}
	effect.DrawCardMultiplier = 2
	if err := effect.ValidateProgram(); err == nil {
		t.Fatal("the exemption admitted a mixed legacy/program action")
	}
}
