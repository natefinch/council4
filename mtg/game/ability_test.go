package game

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
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
			if !slices.Equal(tt.ability.KeywordKinds(), []Keyword{tt.keyword}) {
				t.Fatalf("keywords = %+v, want [%v]", tt.ability.KeywordKinds(), tt.keyword)
			}

			withoutTemplateFields := tt.ability
			withoutTemplateFields.Kind = 0
			withoutTemplateFields.Text = ""
			withoutTemplateFields.Body = nil
			withoutTemplateFields.KeywordAbilities = nil
			if !reflect.DeepEqual(withoutTemplateFields, AbilityDef{}) {
				t.Fatalf("ability has extra fields: %+v", withoutTemplateFields)
			}
		})
	}
}

func TestEternalizeAbilityBuildsKeywordActivation(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.G}
	ability := EternalizeAbility(manaCost, types.Snake, types.Druid)
	manaCost[0] = cost.O(9)

	if ability.Kind != ActivatedAbility || !slices.Equal(ability.KeywordKinds(), []Keyword{Eternalize}) {
		t.Fatalf("ability kind/keywords = %v/%+v, want Eternalize activated ability", ability.Kind, ability.KeywordKinds())
	}
	if !ability.IsActivated() {
		t.Fatal("eternalize body is not an activated body")
	}
	if ability.ZoneOfFunction != ZoneGraveyard || ability.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want graveyard sorcery", ability.ZoneOfFunction, ability.Timing)
	}
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, []cost.Symbol{cost.O(2), cost.G}) {
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

func TestAbilityBodyAccessorsPreferBody(t *testing.T) {
	ability := AbilityDef{
		Kind: SpellAbility,
		Body: TriggeredAbilityBody{
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield},
			},
			Content: PlainAbilityContent{Sequence: []Effect{{Type: EffectDraw}}},
		},
	}

	triggered, ok := ability.TriggeredBody()
	if !ok {
		t.Fatal("TriggeredBody returned false")
	}
	if triggered.Trigger.Pattern.Event != EventPermanentEnteredBattlefield {
		t.Fatalf("trigger event = %v, want permanent entered", triggered.Trigger.Pattern.Event)
	}
	if !ability.IsTriggered() || ability.IsSpell() {
		t.Fatalf("body classification mismatch: triggered=%v spell=%v", ability.IsTriggered(), ability.IsSpell())
	}
	if ability.EffectiveKind() != TriggeredAbility {
		t.Fatalf("effective kind = %v, want TriggeredAbility", ability.EffectiveKind())
	}
}

func TestLegacyAbilityBodyViews(t *testing.T) {
	ability := AbilityDef{
		Kind: ActivatedAbility,
		Targets: []TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
		}},
		Effects: []Effect{{Type: EffectDamage, TargetIndex: 0}},
	}

	body, ok := ability.ActivatedBody()
	if !ok {
		t.Fatal("ActivatedBody returned false")
	}
	content, ok := body.Content.(PlainAbilityContent)
	if !ok {
		t.Fatalf("content = %T, want PlainAbilityContent", body.Content)
	}
	if len(content.Targets) != 1 || len(content.Sequence) != 1 {
		t.Fatalf("content = %+v, want one target and one effect", content)
	}
}

func TestAbilityFieldAccessorsPreferBody(t *testing.T) {
	activationCondition := opt.Val(Condition{ControllerHasMaxSpeed: true})
	ability := AbilityDef{
		Kind:                ActivatedAbility,
		ZoneOfFunction:      ZoneBattlefield,
		Timing:              NoTimingRestriction,
		ActivationCondition: opt.Val(Condition{SourceNotMonstrous: true}),
		LoyaltyCost:         1,
		Body: LoyaltyAbilityBody{
			LoyaltyCost:         -2,
			ActivationCondition: activationCondition,
		},
	}

	if ability.FunctionZone() != ZoneBattlefield {
		t.Fatalf("function zone = %v, want legacy battlefield fallback", ability.FunctionZone())
	}
	if ability.TimingRestriction() != NoTimingRestriction {
		t.Fatalf("timing = %v, want legacy no restriction fallback", ability.TimingRestriction())
	}
	if ability.LoyaltyCostValue() != -2 {
		t.Fatalf("loyalty cost = %v, want body value", ability.LoyaltyCostValue())
	}
	if got := ability.ActivationConditionValue(); !got.Exists || !got.Val.ControllerHasMaxSpeed {
		t.Fatalf("activation condition = %+v, want body condition", got)
	}
}
