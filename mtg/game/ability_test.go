package game

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestSimpleKeywordStaticBodyTemplates(t *testing.T) {
	tests := []struct {
		name    string
		body    StaticAbilityBody
		keyword Keyword
	}{
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
	if !ActivatedBodyEternalize(body) {
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

func TestBodyAccessors(t *testing.T) {
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1}}
	activationCondition := opt.Val(Condition{SourceNotMonstrous: true})
	body := ActivatedAbilityBody{
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

	loyalty := LoyaltyAbilityBody{LoyaltyCost: -2}
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

	modal := ModalAbilityContent{
		Modes:    []Mode{{Text: "First"}, {Text: "Second"}},
		MinModes: 1,
		MaxModes: 1,
	}
	if !modal.IsModal() {
		t.Fatal("multiple modes were treated as non-modal")
	}
}

func TestKeywordBodyHelpers(t *testing.T) {
	wardBody := TriggeredAbilityBody{
		KeywordAbilities: []KeywordAbility{WardKeyword{Cost: cost.Mana{cost.O(2)}}},
	}
	if wardCost, ok := BodyWardCost(wardBody); !ok || !slices.Equal(wardCost, cost.Mana{cost.O(2)}) {
		t.Fatalf("BodyWardCost = %+v/%v, want {2}/true", wardCost, ok)
	}

	madnessBody := TriggeredAbilityBody{
		KeywordAbilities: []KeywordAbility{MadnessKeyword{Cost: cost.Mana{cost.B}}},
	}
	if manaCost, ok := BodyMadnessCost(madnessBody); !ok || !slices.Equal(manaCost, cost.Mana{cost.B}) {
		t.Fatalf("BodyMadnessCost = %+v/%v, want {B}/true", manaCost, ok)
	}

	activated := ActivatedAbilityBody{
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

	staticBody := StaticAbilityBody{
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
