package game

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestSimpleKeywordAbilityTemplates(t *testing.T) {
	tests := []struct {
		name    string
		ability AbilityDef
		keyword Keyword
	}{
		{name: "DeathtouchAbility", ability: DeathtouchAbility, keyword: Deathtouch},
		{name: "DefenderAbility", ability: DefenderAbility, keyword: Defender},
		{name: "DoubleStrikeAbility", ability: DoubleStrikeAbility, keyword: DoubleStrike},
		{name: "FirstStrikeAbility", ability: FirstStrikeAbility, keyword: FirstStrike},
		{name: "FlashAbility", ability: FlashAbility, keyword: Flash},
		{name: "FlyingAbility", ability: FlyingAbility, keyword: Flying},
		{name: "HasteAbility", ability: HasteAbility, keyword: Haste},
		{name: "HexproofAbility", ability: HexproofAbility, keyword: Hexproof},
		{name: "IndestructibleAbility", ability: IndestructibleAbility, keyword: Indestructible},
		{name: "LifelinkAbility", ability: LifelinkAbility, keyword: Lifelink},
		{name: "MenaceAbility", ability: MenaceAbility, keyword: Menace},
		{name: "ReachAbility", ability: ReachAbility, keyword: Reach},
		{name: "ShroudAbility", ability: ShroudAbility, keyword: Shroud},
		{name: "TrampleAbility", ability: TrampleAbility, keyword: Trample},
		{name: "VigilanceAbility", ability: VigilanceAbility, keyword: Vigilance},
		{name: "SplitSecondAbility", ability: SplitSecondAbility, keyword: SplitSecond},
		{name: "ConvokeAbility", ability: ConvokeAbility, keyword: Convoke},
		{name: "DelveAbility", ability: DelveAbility, keyword: Delve},
		{name: "StormAbility", ability: StormAbility, keyword: Storm},
		{name: "CascadeAbility", ability: CascadeAbility, keyword: Cascade},
		{name: "ProwessAbility", ability: ProwessAbility, keyword: Prowess},
		{name: "ImproviseAbility", ability: ImproviseAbility, keyword: Improvise},
		{name: "UndyingAbility", ability: UndyingAbility, keyword: Undying},
		{name: "PersistAbility", ability: PersistAbility, keyword: Persist},
		{name: "WitherAbility", ability: WitherAbility, keyword: Wither},
		{name: "InfectAbility", ability: InfectAbility, keyword: Infect},
		{name: "ExaltedAbility", ability: ExaltedAbility, keyword: Exalted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ability.Kind != StaticAbility {
				t.Fatalf("kind = %v, want StaticAbility", tt.ability.Kind)
			}
			if tt.ability.Text == "" {
				t.Fatal("text is empty")
			}
			if !slices.Equal(tt.ability.Keywords, []Keyword{tt.keyword}) {
				t.Fatalf("keywords = %+v, want [%v]", tt.ability.Keywords, tt.keyword)
			}

			withoutTemplateFields := tt.ability
			withoutTemplateFields.Kind = 0
			withoutTemplateFields.Text = ""
			withoutTemplateFields.Keywords = nil
			if !reflect.DeepEqual(withoutTemplateFields, AbilityDef{}) {
				t.Fatalf("ability has extra fields: %+v", withoutTemplateFields)
			}
		})
	}
}

func TestEternalizeAbilityBuildsKeywordActivation(t *testing.T) {
	cost := mana.Cost{mana.GenericMana(2), mana.G}
	ability := EternalizeAbility(cost, types.Snake, types.Druid)
	cost[0] = mana.GenericMana(9)

	if ability.Kind != ActivatedAbility || !slices.Equal(ability.Keywords, []Keyword{Eternalize}) {
		t.Fatalf("ability kind/keywords = %v/%+v, want Eternalize activated ability", ability.Kind, ability.Keywords)
	}
	if ability.ZoneOfFunction != ZoneGraveyard || ability.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want graveyard sorcery", ability.ZoneOfFunction, ability.Timing)
	}
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, []mana.Symbol{mana.GenericMana(2), mana.G}) {
		t.Fatalf("mana cost = %+v, want copied eternalize cost", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != AdditionalCostExileSource {
		t.Fatalf("additional costs = %+v, want source exile", ability.AdditionalCosts)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Type != EffectCreateToken || !ability.Effects[0].TokenCopy.Exists {
		t.Fatalf("effects = %+v, want create token-copy effect", ability.Effects)
	}
	spec := ability.Effects[0].TokenCopy.Val
	if spec.Source != TokenCopySourceSourceCard || !spec.NoManaCost {
		t.Fatalf("token copy source/no-cost = %v/%v, want source card with no mana cost", spec.Source, spec.NoManaCost)
	}
	if !slices.Equal(spec.SetColors, []color.Color{color.Black}) || !slices.Equal(spec.SetSubtypes, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token colors/subtypes = %+v/%+v, want black Zombie Snake Druid", spec.SetColors, spec.SetSubtypes)
	}
	if !spec.SetPower.Exists || spec.SetPower.Val.Value != 4 || !spec.SetToughness.Exists || spec.SetToughness.Val.Value != 4 {
		t.Fatalf("token P/T = %+v/%+v, want 4/4", spec.SetPower, spec.SetToughness)
	}
}
