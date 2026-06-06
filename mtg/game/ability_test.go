package game

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

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
			if tt.ability.EffectiveKind() != StaticAbility {
				t.Fatalf("effective kind = %v, want StaticAbility", tt.ability.EffectiveKind())
			}
			if tt.ability.Text != "" {
				t.Fatal("text should be empty on body-only template")
			}
			if !slices.Equal(tt.ability.KeywordKinds(), []Keyword{tt.keyword}) {
				t.Fatalf("keywords = %+v, want [%v]", tt.ability.KeywordKinds(), tt.keyword)
			}

			withoutTemplateFields := tt.ability
			withoutTemplateFields.Body = nil
			if !reflect.DeepEqual(withoutTemplateFields, AbilityDef{}) {
				t.Fatalf("ability has extra fields beyond Body: %+v", withoutTemplateFields)
			}
		})
	}
}

func TestEternalizeAbilityBuildsKeywordActivation(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.G}
	ability := EternalizeAbility(manaCost, types.Snake, types.Druid)
	manaCost[0] = cost.O(9)

	if ability.EffectiveKind() != ActivatedAbility || !slices.Equal(ability.KeywordKinds(), []Keyword{Eternalize}) {
		t.Fatalf("effective kind/keywords = %v/%+v, want Eternalize activated ability", ability.EffectiveKind(), ability.KeywordKinds())
	}
	if !ability.IsActivated() {
		t.Fatal("eternalize body is not an activated body")
	}
	body, ok := ability.ActivatedBody()
	if !ok {
		t.Fatal("ActivatedBody returned false for eternalize ability")
	}
	if body.ZoneOfFunction != zone.Graveyard || body.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want graveyard sorcery", body.ZoneOfFunction, body.Timing)
	}
	if !body.ManaCost.Exists || !slices.Equal(body.ManaCost.Val, []cost.Symbol{cost.O(2), cost.G}) {
		t.Fatalf("mana cost = %+v, want copied eternalize cost", body.ManaCost)
	}
	if len(body.AdditionalCosts) != 1 || body.AdditionalCosts[0].Kind != cost.AdditionalExileSource {
		t.Fatalf("additional costs = %+v, want source exile", body.AdditionalCosts)
	}
	content, ok := body.Content.(PlainAbilityContent)
	if !ok {
		t.Fatalf("body content = %T, want PlainAbilityContent", body.Content)
	}
	if len(content.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one create-token instruction", content.Sequence)
	}
	prim, ok := content.Sequence[0].Primitive.(CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want CreateToken", content.Sequence[0].Primitive)
	}
	spec, ok := prim.Source.TokenCopy()
	if !ok {
		t.Fatalf("token source = %+v, want token-copy source", prim.Source)
	}
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
			Content: PlainAbilityContent{LegacyEffects: []Effect{{Type: EffectDraw}}},
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
	if len(content.Targets) != 1 || len(content.LegacyEffects) != 1 {
		t.Fatalf("content = %+v, want one target and one effect", content)
	}
}

func TestAbilityFieldAccessorsPreferBody(t *testing.T) {
	activationCondition := opt.Val(Condition{ControllerHasMaxSpeed: true})
	ability := AbilityDef{
		Kind:                ActivatedAbility,
		ZoneOfFunction:      zone.Battlefield,
		Timing:              NoTimingRestriction,
		ActivationCondition: opt.Val(Condition{SourceNotMonstrous: true}),
		LoyaltyCost:         1,
		Body: LoyaltyAbilityBody{
			LoyaltyCost:         -2,
			ActivationCondition: activationCondition,
		},
	}

	if ability.FunctionZone() != zone.Battlefield {
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

// TestBodyOnlySpellAbilityBodyLowers verifies SpellAbilityBody lowers to flat fields.
func TestBodyOnlySpellAbilityBodyLowers(t *testing.T) {
	src := AbilityDef{Body: SpellAbilityBody{
		Text:    "Draw two cards.",
		Content: PlainAbilityContent{LegacyEffects: []Effect{{Type: EffectDraw, Amount: 2}}},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalTap, Text: "Tap"},
		},
		AlternativeCosts: []cost.Alternative{{Label: "Overload"}},
	}}
	got := src.WithBody()
	if got.Kind != SpellAbility {
		t.Fatalf("Kind = %v, want SpellAbility", got.Kind)
	}
	if got.Text != "Draw two cards." {
		t.Fatalf("Text = %q, want spell text", got.Text)
	}
	if len(got.AdditionalCosts) != 1 || got.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts = %+v, want tap", got.AdditionalCosts)
	}
	if len(got.AlternativeCosts) != 1 {
		t.Fatalf("AlternativeCosts = %+v, want one", got.AlternativeCosts)
	}
	if len(got.Effects) != 1 || got.Effects[0].Type != EffectDraw {
		t.Fatalf("Effects = %+v, want draw", got.Effects)
	}
}

// TestBodyOnlyActivatedEquipAbilityLowers verifies ActivatedAbilityBody with equip keyword lowers to flat fields.
func TestBodyOnlyActivatedEquipAbilityLowers(t *testing.T) {
	equipCost := cost.Mana{cost.O(2)}
	src := AbilityDef{Body: ActivatedAbilityBody{
		Text:     "Equip {2}",
		ManaCost: opt.Val(equipCost),
		Timing:   SorceryOnly,
		Content: PlainAbilityContent{
			Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
		},
		KeywordAbilities: []KeywordAbility{EquipKeyword{Cost: equipCost}},
	}}
	got := src.WithBody()
	if got.Kind != ActivatedAbility {
		t.Fatalf("Kind = %v, want ActivatedAbility", got.Kind)
	}
	if !got.IsActivated() {
		t.Fatal("IsActivated() should be true")
	}
	if got.Timing != SorceryOnly {
		t.Fatalf("Timing = %v, want SorceryOnly", got.Timing)
	}
	if len(got.Targets) != 1 {
		t.Fatalf("Targets = %+v, want one target", got.Targets)
	}
	if !got.HasKeyword(Equip) {
		t.Fatal("HasKeyword(Equip) should be true after lowering")
	}
	if len(got.KeywordAbilities) != 1 {
		t.Fatalf("KeywordAbilities = %+v, want equip keyword", got.KeywordAbilities)
	}
}

// TestBodyOnlyManaAbilityBodyPlainLowers verifies plain ManaAbilityBody lowers to flat fields.
func TestBodyOnlyManaAbilityBodyPlainLowers(t *testing.T) {
	src := AbilityDef{Body: ManaAbilityBody{
		Text:            "{T}: Add {G}.",
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
		Content: PlainAbilityContent{LegacyEffects: []Effect{
			{Type: EffectAddMana, Amount: 1},
		}},
	}}
	got := src.WithBody()
	if got.Kind != ActivatedAbility {
		t.Fatalf("Kind = %v, want ActivatedAbility", got.Kind)
	}
	if !got.IsManaAbility {
		t.Fatal("IsManaAbility should be true")
	}
	if !got.IsMana() {
		t.Fatal("IsMana() should be true")
	}
	if len(got.Effects) != 1 || got.Effects[0].Type != EffectAddMana {
		t.Fatalf("Effects = %+v, want AddMana", got.Effects)
	}
}

// TestBodyOnlyManaAbilityBodyModalLowers verifies modal ManaAbilityBody lowers to flat modes.
func TestBodyOnlyManaAbilityBodyModalLowers(t *testing.T) {
	src := AbilityDef{Body: ManaAbilityBody{
		Text: "{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.",
		Content: ModalAbilityContent{
			Modes: []Mode{
				{Text: "Add {R}{R}.", LegacyEffects: []Effect{{Type: EffectAddMana}, {Type: EffectAddMana}}},
				{Text: "Add {G}{G}.", LegacyEffects: []Effect{{Type: EffectAddMana}, {Type: EffectAddMana}}},
			},
		},
	}}
	got := src.WithBody()
	if got.Kind != ActivatedAbility || !got.IsManaAbility {
		t.Fatalf("Kind/IsMana = %v/%v, want ActivatedAbility mana", got.Kind, got.IsManaAbility)
	}
	if len(got.Modes) != 2 {
		t.Fatalf("Modes = %v, want 2 modes for modal mana", len(got.Modes))
	}
}

// TestBodyOnlyLoyaltyAbilityBodyLowers verifies LoyaltyAbilityBody lowers to flat fields.
func TestBodyOnlyLoyaltyAbilityBodyLowers(t *testing.T) {
	src := AbilityDef{Body: LoyaltyAbilityBody{
		Text:        "-2: Fight.",
		LoyaltyCost: -2,
		Content: PlainAbilityContent{
			Targets:       []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
			LegacyEffects: []Effect{{Type: EffectFight}},
		},
	}}
	got := src.WithBody()
	if got.Kind != ActivatedAbility {
		t.Fatalf("Kind = %v, want ActivatedAbility", got.Kind)
	}
	if !got.IsLoyaltyAbility {
		t.Fatal("IsLoyaltyAbility should be true")
	}
	if !got.IsLoyalty() {
		t.Fatal("IsLoyalty() should be true")
	}
	if got.LoyaltyCost != -2 {
		t.Fatalf("LoyaltyCost = %v, want -2", got.LoyaltyCost)
	}
	if len(got.Targets) != 1 || len(got.Effects) != 1 {
		t.Fatalf("Targets/Effects = %v/%v, want 1/1", len(got.Targets), len(got.Effects))
	}
}

// TestBodyOnlyTriggeredAbilityBodyLowers verifies TriggeredAbilityBody lowers to flat fields.
func TestBodyOnlyTriggeredAbilityBodyLowers(t *testing.T) {
	src := AbilityDef{Body: TriggeredAbilityBody{
		Text: "Whenever this enters, draw a card.",
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
		},
		Optional:           true,
		MaxTriggersPerTurn: 1,
		Content: PlainAbilityContent{LegacyEffects: []Effect{
			{Type: EffectDraw, Amount: 1},
		}},
	}}
	got := src.WithBody()
	if got.Kind != TriggeredAbility {
		t.Fatalf("Kind = %v, want TriggeredAbility", got.Kind)
	}
	if !got.IsTriggered() {
		t.Fatal("IsTriggered() should be true")
	}
	if !got.Trigger.Exists || got.Trigger.Val.Pattern.Event != EventPermanentEnteredBattlefield {
		t.Fatalf("Trigger = %+v, want ETB trigger", got.Trigger)
	}
	if !got.Optional {
		t.Fatal("Optional should be true")
	}
	if got.MaxTriggersPerTurn != 1 {
		t.Fatalf("MaxTriggersPerTurn = %v, want 1", got.MaxTriggersPerTurn)
	}
	if len(got.Effects) != 1 || got.Effects[0].Type != EffectDraw {
		t.Fatalf("Effects = %+v, want draw", got.Effects)
	}
}

// TestBodyOnlyStaticAbilityBodyLowers verifies StaticAbilityBody lowers to flat fields.
func TestBodyOnlyStaticAbilityBodyLowers(t *testing.T) {
	activationCond := opt.Val(Condition{ControllerHasMaxSpeed: true})
	src := AbilityDef{Body: StaticAbilityBody{
		Text:      "Flying",
		Condition: activationCond,
		KeywordAbilities: []KeywordAbility{
			SimpleKeyword{Kind: Flying},
		},
		LegacyEffects: []Effect{{Type: EffectApplyContinuous}},
	}}
	got := src.WithBody()
	if got.Kind != StaticAbility {
		t.Fatalf("Kind = %v, want StaticAbility", got.Kind)
	}
	if !got.IsStatic() {
		t.Fatal("IsStatic() should be true")
	}
	if !got.Condition.Exists || !got.Condition.Val.ControllerHasMaxSpeed {
		t.Fatalf("Condition = %+v, want max-speed condition", got.Condition)
	}
	if !got.HasKeyword(Flying) {
		t.Fatal("HasKeyword(Flying) should be true after lowering")
	}
	if len(got.KeywordAbilities) != 1 {
		t.Fatalf("KeywordAbilities = %+v, want flying keyword", got.KeywordAbilities)
	}
	if len(got.Effects) != 1 {
		t.Fatalf("Effects = %+v, want one effect", got.Effects)
	}
}
