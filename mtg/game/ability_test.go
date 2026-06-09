package game

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestSimpleKeywordStaticBodyTemplates(t *testing.T) {
	tests := []struct {
		name    string
		body    StaticAbility
		keyword Keyword
	}{
		{name: "DevoidStaticBody", body: DevoidStaticBody, keyword: Devoid},
		{name: "DeathtouchStaticBody", body: DeathtouchStaticBody, keyword: Deathtouch},
		{name: "FlashStaticBody", body: FlashStaticBody, keyword: Flash},
		{name: "FlyingStaticBody", body: FlyingStaticBody, keyword: Flying},
		{name: "HexproofStaticBody", body: HexproofStaticBody, keyword: Hexproof},
		{name: "TrampleStaticBody", body: TrampleStaticBody, keyword: Trample},
		{name: "ExaltedStaticBody", body: ExaltedStaticBody, keyword: Exalted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.body.Text == "" {
				t.Fatal("template text should not be empty")
			}
			if !BodyHasKeyword(tt.body, tt.keyword) {
				t.Fatalf("BodyHasKeyword(%v) = false", tt.keyword)
			}
			kas := BodyKeywordAbilities(tt.body)
			if len(kas) != 1 || KeywordAbilityKind(kas[0]) != tt.keyword {
				t.Fatalf("keywords = %+v, want [%v]", kas, tt.keyword)
			}
		})
	}
}

func TestEternalizeActivatedBodyBuildsKeywordActivation(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.G}
	body := EternalizeActivatedBody(manaCost, types.Snake, types.Druid)
	manaCost[0] = cost.O(9)

	if body.ZoneOfFunction != zone.Graveyard || body.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want graveyard sorcery", body.ZoneOfFunction, body.Timing)
	}
	if !body.ManaCost.Exists || !slices.Equal(body.ManaCost.Val, []cost.Symbol{cost.O(2), cost.G}) {
		t.Fatalf("mana cost = %+v, want copied eternalize cost", body.ManaCost)
	}
	if len(body.AdditionalCosts) != 1 || body.AdditionalCosts[0].Kind != cost.AdditionalExileSource {
		t.Fatalf("additional costs = %+v, want source exile", body.AdditionalCosts)
	}
	if !ActivatedBodyEternalize(&body) {
		t.Fatal("ActivatedBodyEternalize() = false")
	}
	if body.Content.IsModal() || len(body.Content.Modes) != 1 {
		t.Fatalf("body content = %+v, want one non-modal mode", body.Content)
	}
	sequence := body.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %+v, want one create-token instruction", sequence)
	}
	prim, ok := sequence[0].Primitive.(CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want CreateToken", sequence[0].Primitive)
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

func TestCyclingActivatedAbilityBuildsCompleteMechanic(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.G}
	ability := CyclingActivatedAbility(manaCost)
	manaCost[0] = cost.O(9)

	if ability.Text != "Cycling {2}{G}" {
		t.Fatalf("text = %q, want %q", ability.Text, "Cycling {2}{G}")
	}
	if ability.ZoneOfFunction != zone.Hand {
		t.Fatalf("zone = %v, want hand", ability.ZoneOfFunction)
	}
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, []cost.Symbol{cost.O(2), cost.G}) {
		t.Fatalf("mana cost = %+v, want copied cycling cost", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard ||
		ability.AdditionalCosts[0].Source != zone.Hand ||
		ability.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("additional costs = %+v, want discard this card from hand", ability.AdditionalCosts)
	}
	keywordCost, ok := ActivatedBodyCyclingCost(&ability)
	if !ok || !slices.Equal(keywordCost, []cost.Symbol{cost.O(2), cost.G}) {
		t.Fatalf("cycling keyword cost = %v, %v; want copied {2}{G}", keywordCost, ok)
	}
	content := BodyContent(ability)
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %+v, want one non-modal instruction", content)
	}
	draw, ok := content.Modes[0].Sequence[0].Primitive.(Draw)
	if !ok || draw.Amount.Value() != 1 || draw.Player != ControllerReference() {
		t.Fatalf("instruction = %+v, want controller draws one", content.Modes[0].Sequence[0])
	}
}

func TestEquipActivatedAbilityBuildsCompleteMechanic(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.R}
	ability := EquipActivatedAbility(manaCost)
	manaCost[0] = cost.O(9)

	if ability.Text != "Equip {2}{R}" {
		t.Fatalf("text = %q, want %q", ability.Text, "Equip {2}{R}")
	}
	if ability.ZoneOfFunction != zone.Battlefield || ability.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want battlefield/sorcery", ability.ZoneOfFunction, ability.Timing)
	}
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, []cost.Symbol{cost.O(2), cost.R}) {
		t.Fatalf("mana cost = %+v, want copied equip cost", ability.ManaCost)
	}
	keywordCost, ok := ActivatedBodyEquipCost(&ability)
	if !ok || !slices.Equal(keywordCost, []cost.Symbol{cost.O(2), cost.R}) {
		t.Fatalf("equip keyword cost = %v, %v; want copied {2}{R}", keywordCost, ok)
	}
	targets := BodyTargets(ability)
	if len(targets) != 1 ||
		targets[0].MinTargets != 1 ||
		targets[0].MaxTargets != 1 ||
		targets[0].Allow != TargetAllowPermanent ||
		!slices.Equal(targets[0].Predicate.PermanentTypes, []types.Card{types.Creature}) ||
		targets[0].Predicate.Controller != ControllerYou {
		t.Fatalf("targets = %+v, want one creature you control", targets)
	}
}

func TestCantBeCounteredStaticBodyBuildsCompleteMechanic(t *testing.T) {
	if CantBeCounteredStaticBody.Text != "This spell can't be countered." {
		t.Fatalf("text = %q", CantBeCounteredStaticBody.Text)
	}
	if CantBeCounteredStaticBody.ZoneOfFunction != zone.Stack {
		t.Fatalf("zone = %v, want stack", CantBeCounteredStaticBody.ZoneOfFunction)
	}
	if len(CantBeCounteredStaticBody.RuleEffects) != 1 {
		t.Fatalf("rule effects = %+v", CantBeCounteredStaticBody.RuleEffects)
	}
	effect := CantBeCounteredStaticBody.RuleEffects[0]
	if effect.Kind != RuleEffectCantBeCountered || !effect.AffectedSource {
		t.Fatalf("rule effect = %+v", effect)
	}
}

func TestWardStaticAbilityBuildsCompleteMechanic(t *testing.T) {
	manaCost := cost.Mana{cost.O(2), cost.U}
	ability := WardStaticAbility(manaCost)
	manaCost[0] = cost.O(9)

	if ability.Text != "Ward {2}{U}" {
		t.Fatalf("text = %q, want %q", ability.Text, "Ward {2}{U}")
	}
	keywordCost, ok := StaticBodyWardCost(&ability)
	if !ok || !slices.Equal(keywordCost, []cost.Symbol{cost.O(2), cost.U}) {
		t.Fatalf("ward keyword cost = %v, %v; want copied {2}{U}", keywordCost, ok)
	}
}

func TestEnchantStaticAbilityBuildsCompleteMechanic(t *testing.T) {
	target := TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "creature",
		Allow:      TargetAllowPermanent,
		Predicate: TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
		},
	}
	ability := EnchantStaticAbility(&target)
	target.Predicate.PermanentTypes[0] = types.Land

	if ability.Text != "Enchant creature" {
		t.Fatalf("text = %q, want %q", ability.Text, "Enchant creature")
	}
	enchantTarget, ok := StaticBodyEnchantTarget(&ability)
	if !ok ||
		enchantTarget.MinTargets != 1 ||
		enchantTarget.MaxTargets != 1 ||
		enchantTarget.Allow != TargetAllowPermanent ||
		!slices.Equal(enchantTarget.Predicate.PermanentTypes, []types.Card{types.Creature}) {
		t.Fatalf("enchant target = %+v, %v; want one creature", enchantTarget, ok)
	}
}

func TestProtectionFromColorsStaticAbilityBuildsCompleteMechanic(t *testing.T) {
	colors := []color.Color{color.Red}
	ability := ProtectionFromColorsStaticAbility(colors...)
	colors[0] = color.Blue

	if ability.Text != "Protection from red" {
		t.Fatalf("text = %q, want %q", ability.Text, "Protection from red")
	}
	if protected := StaticBodyProtectionColors(&ability); !slices.Equal(protected, []color.Color{color.Red}) {
		t.Fatalf("protection colors = %v, want red", protected)
	}

	multiple := ProtectionFromColorsStaticAbility(color.Black, color.Red)
	if multiple.Text != "Protection from black and from red" {
		t.Fatalf("multiple text = %q, want %q", multiple.Text, "Protection from black and from red")
	}
}

func TestCantBlockStaticBodyBuildsCompleteMechanic(t *testing.T) {
	if CantBlockStaticBody.Text != "This creature can't block." {
		t.Fatalf("text = %q", CantBlockStaticBody.Text)
	}
	if len(CantBlockStaticBody.RuleEffects) != 1 {
		t.Fatalf("rule effects = %+v", CantBlockStaticBody.RuleEffects)
	}
	effect := CantBlockStaticBody.RuleEffects[0]
	if effect.Kind != RuleEffectCantBlock || !effect.AffectedSource {
		t.Fatalf("rule effect = %+v", effect)
	}
}

func TestTapManaAbilityBuildsCompleteMechanic(t *testing.T) {
	ability := TapManaAbility(mana.G)

	if ability.Text != "{T}: Add {G}." {
		t.Fatalf("text = %q, want %q", ability.Text, "{T}: Add {G}.")
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0] != cost.T {
		t.Fatalf("additional costs = %+v, want tap", ability.AdditionalCosts)
	}
	content := BodyContent(ability)
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %+v, want one non-modal instruction", content)
	}
	add, ok := content.Modes[0].Sequence[0].Primitive.(AddMana)
	if !ok || add.Amount.Value() != 1 || add.ManaColor != mana.G || add.ChoiceFrom != "" {
		t.Fatalf("instruction = %+v, want add one green mana", content.Modes[0].Sequence[0])
	}
}

func TestTapManaAbilityUsesOracleColorlessSymbol(t *testing.T) {
	ability := TapManaAbility(mana.C)
	if ability.Text != "{T}: Add {C}." {
		t.Fatalf("text = %q, want %q", ability.Text, "{T}: Add {C}.")
	}
}

func TestTapManaChoiceAbilityBuildsCompleteMechanic(t *testing.T) {
	colors := []mana.Color{mana.B, mana.R}
	ability := TapManaChoiceAbility(colors...)
	colors[0] = mana.G

	if ability.Text != "{T}: Add {B} or {R}." {
		t.Fatalf("text = %q, want %q", ability.Text, "{T}: Add {B} or {R}.")
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0] != cost.T {
		t.Fatalf("additional costs = %+v, want tap", ability.AdditionalCosts)
	}
	content := BodyContent(ability)
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 2 {
		t.Fatalf("content = %+v, want choose then add", content)
	}
	choose, ok := content.Modes[0].Sequence[0].Primitive.(Choose)
	if !ok ||
		choose.Choice.Kind != ResolutionChoiceMana ||
		!slices.Equal(choose.Choice.Colors, []mana.Color{mana.B, mana.R}) ||
		choose.PublishChoice == "" {
		t.Fatalf("first instruction = %+v, want copied black/red mana choice", content.Modes[0].Sequence[0])
	}
	add, ok := content.Modes[0].Sequence[1].Primitive.(AddMana)
	if !ok || add.Amount.Value() != 1 || add.ManaColor != "" || add.ChoiceFrom != choose.PublishChoice {
		t.Fatalf("second instruction = %+v, want one mana from published choice", content.Modes[0].Sequence[1])
	}
}

func TestTapManaChoiceAbilitySupportsColorlessMana(t *testing.T) {
	ability := TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.C)

	if ability.Text != "{T}: Add {W}, {U}, {B}, or {C}." {
		t.Fatalf("text = %q, want explicit mana symbols", ability.Text)
	}
	choose, ok := ability.Content.Modes[0].Sequence[0].Primitive.(Choose)
	if !ok || choose.Choice.Prompt != "Choose a type of mana" {
		t.Fatalf("choice = %+v, want mana-type prompt", choose)
	}
}

func TestBodyAccessors(t *testing.T) {
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1}}
	activationCondition := opt.Val(Condition{SourceNotMonstrous: true})
	body := ActivatedAbility{
		Text:                "Equip {2}",
		ManaCost:            opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts:     cost.Tap,
		AlternativeCosts:    []cost.Alternative{{Label: "Alt"}},
		ZoneOfFunction:      zone.Graveyard,
		Timing:              SorceryOnly,
		ActivationCondition: activationCondition,
		Content:             Mode{Targets: targets}.Ability(),
		KeywordAbilities:    []KeywordAbility{EquipKeyword{Cost: cost.Mana{cost.O(2)}}},
	}

	if BodyFunctionZone(body) != zone.Graveyard {
		t.Fatalf("BodyFunctionZone = %v, want graveyard", BodyFunctionZone(body))
	}
	if BodyTimingRestriction(body) != SorceryOnly {
		t.Fatalf("BodyTimingRestriction = %v, want SorceryOnly", BodyTimingRestriction(body))
	}
	gotCondition := BodyActivationCondition(body)
	if !gotCondition.Exists || !gotCondition.Val.SourceNotMonstrous {
		t.Fatalf("BodyActivationCondition = %+v, want SourceNotMonstrous", gotCondition)
	}
	if !BodyHasKeyword(body, Equip) {
		t.Fatal("BodyHasKeyword(Equip) = false")
	}
	gotTargets := BodyTargets(body)
	if len(gotTargets) != 1 || gotTargets[0].MinTargets != targets[0].MinTargets || gotTargets[0].MaxTargets != targets[0].MaxTargets {
		t.Fatalf("BodyTargets = %+v, want %+v", gotTargets, targets)
	}

	loyalty := LoyaltyAbility{LoyaltyCost: -2}
	if BodyLoyaltyCost(loyalty) != -2 {
		t.Fatalf("BodyLoyaltyCost = %d, want -2", BodyLoyaltyCost(loyalty))
	}
}

func TestModalAbilityContentIsModal(t *testing.T) {
	ordinary := Mode{Text: "Draw a card."}.Ability()
	if ordinary.IsModal() {
		t.Fatal("one required mode was treated as modal")
	}
	if ordinary.MinModes != 1 || ordinary.MaxModes != 1 || len(ordinary.Modes) != 1 {
		t.Fatalf("Mode.Ability() = %+v, want one required mode", ordinary)
	}

	modal := AbilityContent{
		Modes:    []Mode{{Text: "First"}, {Text: "Second"}},
		MinModes: 1,
		MaxModes: 1,
	}
	if !modal.IsModal() {
		t.Fatal("multiple modes were treated as non-modal")
	}
}

func TestKeywordBodyHelpers(t *testing.T) {
	wardBody := &TriggeredAbility{
		KeywordAbilities: []KeywordAbility{WardKeyword{Cost: cost.Mana{cost.O(2)}}},
	}
	if wardCost, ok := BodyWardCost(wardBody); !ok || !slices.Equal(wardCost, cost.Mana{cost.O(2)}) {
		t.Fatalf("BodyWardCost = %+v/%v, want {2}/true", wardCost, ok)
	}

	madnessBody := &TriggeredAbility{
		KeywordAbilities: []KeywordAbility{MadnessKeyword{Cost: cost.Mana{cost.B}}},
	}
	if manaCost, ok := BodyMadnessCost(madnessBody); !ok || !slices.Equal(manaCost, cost.Mana{cost.B}) {
		t.Fatalf("BodyMadnessCost = %+v/%v, want {B}/true", manaCost, ok)
	}

	activated := &ActivatedAbility{
		KeywordAbilities: []KeywordAbility{
			SuspendKeyword{Cost: cost.Mana{cost.U}, TimeCounters: 3},
			MorphKeyword{Cost: cost.Mana{cost.O(3)}},
			DisguiseKeyword{Cost: cost.Mana{cost.O(2), cost.U}},
			KickerKeyword{Cost: cost.Mana{cost.R}},
		},
	}
	if manaCost, counters, ok := ActivatedBodySuspendInfo(activated); !ok || counters != 3 || !slices.Equal(manaCost, cost.Mana{cost.U}) {
		t.Fatalf("ActivatedBodySuspendInfo = %+v/%d/%v, want {U}/3/true", manaCost, counters, ok)
	}
	if manaCost, ok := ActivatedBodyMorphCost(activated); !ok || !slices.Equal(manaCost, cost.Mana{cost.O(3)}) {
		t.Fatalf("ActivatedBodyMorphCost = %+v/%v, want {3}/true", manaCost, ok)
	}
	if manaCost, ok := ActivatedBodyDisguiseCost(activated); !ok || !slices.Equal(manaCost, cost.Mana{cost.O(2), cost.U}) {
		t.Fatalf("ActivatedBodyDisguiseCost = %+v/%v, want {2}{U}/true", manaCost, ok)
	}
	if kicker, ok := ActivatedBodyKicker(activated); !ok || !slices.Equal(kicker.Cost, cost.Mana{cost.R}) {
		t.Fatalf("ActivatedBodyKicker = %+v/%v, want {R}/true", kicker, ok)
	}

	staticBody := &StaticAbility{
		KeywordAbilities: []KeywordAbility{
			EnchantKeyword{Target: TargetSpec{Constraint: "creature"}},
			ProtectionKeyword{FromColors: []color.Color{color.Black, color.Red}},
		},
	}
	target, ok := StaticBodyEnchantTarget(staticBody)
	if !ok || target.Constraint != "creature" {
		t.Fatalf("StaticBodyEnchantTarget = %+v/%v, want creature/true", target, ok)
	}
	if colors := StaticBodyProtectionColors(staticBody); !slices.Equal(colors, []color.Color{color.Black, color.Red}) {
		t.Fatalf("StaticBodyProtectionColors = %+v, want black/red", colors)
	}
}
