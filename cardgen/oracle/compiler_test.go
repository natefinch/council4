package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileActivatedAbility(t *testing.T) {
	t.Parallel()
	source := "{1}{G}, {T}: Target attacking creature you control gets +2/+2 until end of turn."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 2 {
		t.Fatalf("cost = %#v", ability.Cost)
	}

	if ability.Cost.Components[0].Kind != CostMana ||
		ability.Cost.Components[0].Symbol != "{1}{G}" ||
		ability.Cost.Components[1].Kind != CostTap {
		t.Fatalf("cost components = %#v", ability.Cost.Components)
	}
	if len(ability.Targets) != 1 {
		t.Fatalf("targets = %#v", ability.Targets)
	}
	target := ability.Targets[0]
	if target.Selector.Kind != SelectorCreature ||
		target.Selector.Controller != ControllerYou ||
		!target.Selector.Attacking {
		t.Fatalf("target selector = %#v", target.Selector)
	}
	if len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != EffectModifyPT ||
		ability.Effects[0].Duration != DurationUntilEndOfTurn {
		t.Fatalf("effects = %#v", ability.Effects)
	}
	if ability.Effects[0].PowerDelta != (CompiledSignedAmount{Value: 2, Known: true}) ||
		ability.Effects[0].ToughnessDelta != (CompiledSignedAmount{Value: 2, Known: true}) {
		t.Fatalf("power/toughness change = %#v", ability.Effects[0])
	}
}

func TestCompileActivatedAbilityTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want ActivationTimingKind
	}{
		{"sorcery", "{1}: Draw a card. Activate only as a sorcery.", ActivationTimingSorcery},
		{"once per turn", "{1}: Draw a card. Activate only once each turn.", ActivationTimingOncePerTurn},
		{"combat", "{1}: Draw a card. Activate only during combat.", ActivationTimingDuringCombat},
		{"upkeep", "{1}: Draw a card. Activate only during your upkeep.", ActivationTimingDuringUpkeep},
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
			ActivationTimingSorceryOncePerTurn,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.text, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.ActivationTiming != test.want {
				t.Fatalf("activation timing = %v, want %v", ability.ActivationTiming, test.want)
			}
			if got := test.text[ability.ActivationTimingSpan.Start.Offset:ability.ActivationTimingSpan.End.Offset]; got == "" {
				t.Fatal("activation timing span is empty")
			}
			if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectDraw {
				t.Fatalf("effects = %#v, want one draw effect", ability.Effects)
			}
			if len(ability.References) != 0 {
				t.Fatalf("references = %#v, want timing references excluded", ability.References)
			}
		})
	}
}

func TestCompileActivatedAbilityTapPermanentsCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Tap two untapped artifacts you control: Draw a card.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostTapPermanents || component.Object != "two untapped artifacts you control" {
		t.Fatalf("cost component = %#v, want tap-permanents object", component)
	}
}

func TestCompileActivatedAbilityPluralRemoveCounterCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Remove two storage counters from this land: Add {G}.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostRemoveCounter || component.Object != "two storage counters from this land" {
		t.Fatalf("cost component = %#v, want remove-counter object", component)
	}
}

func TestCompileActivatedAbilityEnergyCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Pay {E}{E}: Draw a card.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostEnergy || component.Amount != "2" {
		t.Fatalf("cost component = %#v, want two-energy cost", component)
	}
}

func TestCompileTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "Whenever a creature enters, if it was cast, draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "a creature enters" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if ability.Trigger.Condition == nil || !ability.Trigger.Condition.Intervening {
		t.Fatalf("intervening condition = %#v", ability.Trigger.Condition)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Effects)
	}
}

func TestCompileTriggeredAbilityWithInternalEventComma(t *testing.T) {
	t.Parallel()
	source := "Whenever you cast a noncreature, nonland spell, draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "you cast a noncreature, nonland spell" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Effects)
	}
}

func TestCompileSagaChapterAbility(t *testing.T) {
	t.Parallel()
	source := "II, III — Draw a card."
	compilation, diagnostics := Compile(source, ParseContext{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityChapter || !slices.Equal(ability.Chapters, []int{2, 3}) {
		t.Fatalf("ability = %#v", ability)
	}
	if ability.AbilityWord != "" {
		t.Fatalf("ability word = %q, want empty", ability.AbilityWord)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Effects)
	}
}

func TestCompileOptionalTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, you may draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if !ability.Optional || source[ability.OptionalSpan.Start.Offset:ability.OptionalSpan.End.Offset] != "you may" {
		t.Fatalf("optional ability = %#v", ability)
	}
}

func TestCompileSelfCannotBlockStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't block."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != EffectCantBlock ||
		!ability.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.References) != 1 ||
		ability.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.References)
	}
}

func TestCompileSelfCannotBeBlockedStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't be blocked."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != EffectCantBeBlocked ||
		!ability.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.References) != 1 ||
		ability.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.References)
	}
}

func TestCompileSelfMustAttackStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature attacks each combat if able."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != EffectMustAttack ||
		ability.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.References) != 1 ||
		ability.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.References)
	}
	if len(ability.Conditions) != 0 {
		t.Fatalf("intrinsic if-able text became a separate condition: %#v", ability.Conditions)
	}
}

func TestCompileSelfUncounterableStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This spell can't be countered."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Effects) != 1 ||
		ability.Effects[0].Kind != EffectCantBeCountered ||
		!ability.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.References) != 1 ||
		ability.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.References)
	}
}

func TestCompileReturnToOwnersHand(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Return target creature to its owner's hand.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectReturn {
		t.Fatalf("effects = %#v", ability.Effects)
	}
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Selector.Kind != SelectorCreature ||
		ability.Targets[0].Text != "target creature to its owner's hand" {
		t.Fatalf("targets = %#v", ability.Targets)
	}
	if len(ability.References) != 1 ||
		ability.References[0].Kind != ReferencePronoun ||
		ability.References[0].Text != "its" {
		t.Fatalf("references = %#v", ability.References)
	}
	if len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.Modes) != 0 ||
		ability.Effects[0].Negated ||
		ability.Targets[0].Cardinality.Min != 1 ||
		ability.Targets[0].Cardinality.Max != 1 {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestCompileGraveyardReturnZones(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		text     string
		fromZone zone.Type
		toZone   zone.Type
	}{
		{
			name:     "target card to hand",
			text:     "Return target instant or sorcery card from your graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "target card to library",
			text:     "Put target card from your graveyard on the bottom of your library.",
			fromZone: zone.Graveyard,
			toZone:   zone.Library,
		},
		{
			name:     "opponents graveyard",
			text:     "Return target creature card from an opponent's graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "self to battlefield",
			text:     "Return this card from your graveyard to the battlefield tapped.",
			fromZone: zone.Graveyard,
			toZone:   zone.Battlefield,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(tc.text, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Effects) != 1 {
				t.Fatalf("effects = %#v", ability.Effects)
			}
			effect := ability.Effects[0]
			if effect.FromZone != tc.fromZone || effect.ToZone != tc.toZone {
				t.Fatalf("zones = %v -> %v, want %v -> %v", effect.FromZone, effect.ToZone, tc.fromZone, tc.toZone)
			}
		})
	}
}

func TestCompileResolutionConditionIsNotIntervening(t *testing.T) {
	t.Parallel()
	source := "When this creature dies, draw a card if you control a Forest."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil {
		t.Fatal("missing trigger")
	}
	if ability.Trigger.Condition != nil {
		t.Fatalf("resolution condition became trigger condition: %#v", ability.Trigger.Condition)
	}
	if len(ability.Conditions) != 1 || ability.Conditions[0].Intervening {
		t.Fatalf("conditions = %#v", ability.Conditions)
	}
}

func TestCompileModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one —\n• Destroy target creature.\n• Draw two cards."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Modes) != 2 {
		t.Fatalf("modes = %#v", ability.Modes)
	}
	if ability.Modes[0].Effects[0].Kind != EffectDestroy ||
		len(ability.Modes[0].Targets) != 1 ||
		ability.Modes[1].Effects[0].Kind != EffectDraw {
		t.Fatalf("compiled modes = %#v", ability.Modes)
	}
}

func TestCompileKeywordsAndReminder(t *testing.T) {
	t.Parallel()
	source := "First strike (This creature deals combat damage before other creatures.)\nEquip {2} ({2}: Attach to target creature you control.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := compilation.Abilities[0].Keywords; len(got) != 1 || got[0].Name != "First strike" {
		t.Fatalf("first strike = %#v", got)
	}
	equip := compilation.Abilities[1]
	if len(equip.Keywords) != 1 || equip.Keywords[0].Name != "Equip" ||
		equip.Keywords[0].Parameter != "{2}" {
		t.Fatalf("equip = %#v", equip.Keywords)
	}
	if len(equip.Effects) != 0 || len(equip.Targets) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", equip)
	}
}

func TestCompileDevoidAndReminder(t *testing.T) {
	t.Parallel()
	source := "Devoid (This card has no color.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Keywords) != 1 ||
		ability.Keywords[0].Name != "Devoid" ||
		ability.Keywords[0].Text != "Devoid" {
		t.Fatalf("keywords = %#v", ability.Keywords)
	}
	if len(ability.Effects) != 0 || len(ability.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileReadAheadAndReminder(t *testing.T) {
	t.Parallel()
	source := "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", compilation.Abilities)
	}
	ability := compilation.Abilities[0]
	if len(ability.Keywords) != 1 ||
		ability.Keywords[0].Name != "Read ahead" ||
		ability.Keywords[0].Text != "Read ahead" {
		t.Fatalf("keywords = %#v", ability.Keywords)
	}
	if len(ability.Effects) != 0 || len(ability.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileEnchantKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Enchant creature", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Enchant" ||
		keywords[0].Parameter != "creature" ||
		keywords[0].Text != "Enchant creature" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Enchant creature") {
		t.Fatalf("enchant keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Protection from red", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Protection" ||
		keywords[0].Parameter != "red" ||
		keywords[0].Text != "Protection from red" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Protection from red") {
		t.Fatalf("protection keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordMultipleColors(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Protection from black and from red", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Keywords
	if len(keywords) != 1 ||
		keywords[0].Parameter != "black,red" ||
		keywords[0].Text != "Protection from black and from red" ||
		keywords[0].Span.End.Offset != len("Protection from black and from red") {
		t.Fatalf("protection keyword = %#v", keywords)
	}
}

func TestCompileTargetsAndReferences(t *testing.T) {
	t.Parallel()
	source := "Legolas deals damage to up to one target creature you don't control. It gains trample until end of turn."
	compilation, diagnostics := Compile(source, ParseContext{
		CardName: "Legolas",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Targets) != 1 ||
		ability.Targets[0].Cardinality != (TargetCardinality{Min: 0, Max: 1}) ||
		ability.Targets[0].Selector.Controller != ControllerNotYou {
		t.Fatalf("targets = %#v", ability.Targets)
	}
	if len(ability.References) != 2 ||
		ability.References[0].Kind != ReferenceSelfName ||
		ability.References[1].Kind != ReferencePronoun {
		t.Fatalf("references = %#v", ability.References)
	}
}

func TestCompileExactTargetCardinalityAndPluralSelector(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Tap two target creatures.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	targets := compilation.Abilities[0].Targets
	if len(targets) != 1 ||
		targets[0].Cardinality != (TargetCardinality{Min: 2, Max: 2}) ||
		targets[0].Selector.Kind != SelectorCreature {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestCompileThirdPersonEffects(t *testing.T) {
	t.Parallel()
	tests := map[string]EffectKind{
		"Each opponent discards a card.":        EffectDiscard,
		"Target player draws two cards.":        EffectDraw,
		"Its controller sacrifices a creature.": EffectSacrifice,
		"That player searches their library.":   EffectSearch,
		"That creature transforms.":             EffectTransform,
	}
	for source, want := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Effects
			if len(effects) != 1 || effects[0].Kind != want {
				t.Fatalf("effects = %#v, want %v", effects, want)
			}
		})
	}
}

func TestCompileFixedEffectValues(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		context ParseContext
		kind    EffectKind
		amount  int
		symbol  string
	}{
		"Draw two cards.": {
			context: ParseContext{InstantOrSorcery: true},
			kind:    EffectDraw,
			amount:  2,
		},
		"Shock deals 3 damage to any target.": {
			context: ParseContext{CardName: "Shock", InstantOrSorcery: true},
			kind:    EffectDealDamage,
			amount:  3,
		},
		"{T}: Add {G}.": {
			kind:   EffectAddMana,
			amount: 1,
			symbol: "{G}",
		},
	}

	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			effect := effects[0]
			if effect.Kind != test.kind ||
				!effect.Amount.Known ||
				effect.Amount.Value != test.amount ||
				effect.Symbol != test.symbol {
				t.Fatalf("effect = %#v", effect)
			}
		})
	}
}

func TestCompileDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		context    ParseContext
		kind       DynamicAmountKind
		form       DynamicAmountForm
		multiplier int
		selector   SelectorKind
		controller ControllerKind
		text       string
	}{
		{"Swarm deals damage equal to the number of creatures you control to any target.", ParseContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 1, SelectorCreature, ControllerYou, "equal to the number of creatures you control"},
		{"Swarm deals damage equal to twice the number of lands on the battlefield to any target.", ParseContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 2, SelectorLand, ControllerAny, "equal to twice the number of lands on the battlefield"},
		{"You gain 2 life for each opponent you have.", ParseContext{InstantOrSorcery: true}, DynamicAmountOpponentCount, DynamicAmountForEach, 2, SelectorUnknown, ControllerAny, "for each opponent you have"},
		{"You gain life equal to your life total.", ParseContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to your life total"},
		{"You gain X life, where X is your life total.", ParseContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is your life total"},
		{"When this creature dies, it deals damage equal to its power to any target.", ParseContext{CardName: "Devil"}, DynamicAmountSourcePower, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to its power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Druid's power.", ParseContext{CardName: "Druid"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Druid's power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Fight Bear's power.", ParseContext{CardName: "Fight Bear"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Fight Bear's power"},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := compilation.Abilities[0].Effects[0].Amount
			if amount.DynamicKind != test.kind ||
				amount.DynamicForm != test.form ||
				amount.Multiplier != test.multiplier ||
				amount.Selector.Kind != test.selector ||
				amount.Selector.Controller != test.controller ||
				amount.Text != test.text {
				t.Fatalf("amount = %#v tokens = %#v", amount, compilation.Syntax.Abilities[0].Tokens)
			}
			if test.kind == DynamicAmountSourcePower && amount.ReferenceSpan == (Span{}) {
				t.Fatal("source-power amount has no reference span")
			}
		})
	}
}

func TestCompileWithCyclingTargetSelector(t *testing.T) {
	t.Parallel()
	source := "Return up to two target cards with cycling from your graveyard to your hand."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	target := compilation.Abilities[0].Targets[0]
	if target.Cardinality.Min != 0 || target.Cardinality.Max != 2 {
		t.Fatalf("cardinality = %#v, want up to two", target.Cardinality)
	}
	if target.Selector.Kind != SelectorCard || target.Selector.Keyword != "Cycling" {
		t.Fatalf("selector = %#v, want card with Cycling", target.Selector)
	}
}

func TestCompileDynamicCardCountWithCyclingInGraveyard(t *testing.T) {
	t.Parallel()
	source := "Flare deals X damage to any target, where X is the number of cards with a cycling ability in your graveyard."
	compilation, diagnostics := Compile(source, ParseContext{CardName: "Flare", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	amount := compilation.Abilities[0].Effects[0].Amount
	if amount.DynamicKind != DynamicAmountCount ||
		amount.DynamicForm != DynamicAmountWhereX ||
		amount.Selector.Kind != SelectorCard ||
		amount.Selector.Keyword != "Cycling" ||
		amount.Selector.Zone != zone.Graveyard ||
		amount.Selector.Controller != ControllerYou {
		t.Fatalf("amount = %#v, want count of cards with Cycling in your graveyard", amount)
	}
}

func TestCompileNamedCounterKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kind counter.Kind
	}{
		{"+1/+1", counter.PlusOnePlusOne},
		{"charge", counter.Charge},
		{"first strike", counter.FirstStrike},
		{"poison", counter.Poison},
		{"experience", counter.Experience},
	}
	for _, test := range tests {
		source := "Put a " + test.name + " counter on target permanent."
		compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Effects[0]
		if !effect.CounterKindKnown || effect.CounterKind != test.kind {
			t.Fatalf("%q counter kind = %v, %v", source, effect.CounterKind, effect.CounterKindKnown)
		}
	}

	compilation, diagnostics := Compile(
		"Put a quest counter on target permanent.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("unknown counter diagnostics = %#v", diagnostics)
	}
	if compilation.Abilities[0].Effects[0].CounterKindKnown {
		t.Fatal("unknown counter kind was recognized")
	}
}

func TestCompileEntersWithCounterKind(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"This creature enters with three +1/+1 counters on it.",
		ParseContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Effects[0]
	if effect.Kind != EffectEnterTapped ||
		!effect.CounterKindKnown ||
		effect.CounterKind != counter.PlusOnePlusOne ||
		!effect.Amount.Known ||
		effect.Amount.Value != 3 {
		t.Fatalf("effect = %#v, want fixed +1/+1 ETB counters", effect)
	}
}

func TestCompileNamedCounterKindsRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"stun", "finality"} {
		source := "Put a " + name + " counter on target creature."
		compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Effects[0]
		if effect.CounterKindKnown {
			t.Fatalf("%q counter kind was accepted for placement", source)
		}
	}
}

func TestCompileDynamicEffectAmountsRejectsAmbiguousSubjects(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Swarm deals damage equal to the number of cards in your hand to any target.",
		"Swarm deals damage equal to the number of creatures you control plus one to any target.",
		"You gain 2 life for each opponent and creature.",
		"Swarm deals damage equal to creatures you control to any target.",
		"You gain X life, where X is opponent.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileDynamicEffectAmountsRejectsNumberDisagreement(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw a card for each creatures you control.",
		"Swarm deals damage equal to the number of creature you control to any target.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileEffectAmountsAreClauseLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, []CompiledEffect)
	}{
		{
			name:   "fixed then dynamic effect",
			source: "You gain 2 life, then draw a card for each creature you control.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
		{
			name:   "dynamic then fixed effect",
			source: "Draw a card for each creature you control, then you gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "and separates effects",
			source: "Draw a card for each creature you control and gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "fixed before condition formula",
			source: "You gain 2 life if the number of creatures you control is greater than 3.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "dynamic before condition amount",
			source: "Draw a card for each creature you control unless your life total is 2.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Effects)
		})
	}
}

func assertFixedEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, value int) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			if !effect.Amount.Known ||
				effect.Amount.Value != value ||
				effect.Amount.DynamicKind != DynamicAmountNone {
				t.Fatalf("%v amount = %#v, want fixed %d", kind, effect.Amount, value)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}

func assertDynamicEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, dynamicKind DynamicAmountKind) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			if effect.Amount.Known || effect.Amount.DynamicKind != dynamicKind {
				t.Fatalf("%v amount = %#v, want dynamic %v", kind, effect.Amount, dynamicKind)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}

func TestCompileStaticPTBuffSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source          string
		wantSubject     StaticSubjectKind
		wantSubjectText string
		wantPower       CompiledSignedAmount
		wantToughness   CompiledSignedAmount
	}{
		"enchanted creature": {
			source:          "Enchanted creature gets +2/+2.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Enchanted creature",
			wantPower:       CompiledSignedAmount{Value: 2, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"equipped creature": {
			source:          "Equipped creature gets -3/-1.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Equipped creature",
			wantPower:       CompiledSignedAmount{Value: 3, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true, Negative: true},
		},
		"other creatures you control": {
			source:          "Other creatures you control get +1/+1.",
			wantSubject:     StaticSubjectOtherControlledCreatures,
			wantSubjectText: "Other creatures you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures you control": {
			source:          "Creatures you control get +0/+2.",
			wantSubject:     StaticSubjectControlledCreatures,
			wantSubjectText: "Creatures you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"each wall you control": {
			source:          "Each Wall you control gets +0/+2.",
			wantSubject:     StaticSubjectControlledWalls,
			wantSubjectText: "Each Wall you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"artifacts you control": {
			source:          "Artifacts you control get +1/+1.",
			wantSubject:     StaticSubjectControlledArtifacts,
			wantSubjectText: "Artifacts you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"tokens you control": {
			source:          "Tokens you control get +1/+1.",
			wantSubject:     StaticSubjectControlledTokens,
			wantSubjectText: "Tokens you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures your opponents control": {
			source:          "Creatures your opponents control get -1/-0.",
			wantSubject:     StaticSubjectOpponentControlledCreatures,
			wantSubjectText: "Creatures your opponents control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 0, Known: true, Negative: true},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 {
				t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
			}
			ability := compilation.Abilities[0]
			if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectModifyPT {
				t.Fatalf("effects = %#v", ability.Effects)
			}
			effect := ability.Effects[0]
			if effect.StaticSubject != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", effect.StaticSubject, test.wantSubject)
			}
			if got := test.source[effect.StaticSubjectSpan.Start.Offset:effect.StaticSubjectSpan.End.Offset]; got != test.wantSubjectText {
				t.Fatalf("subject span text = %q, want %q", got, test.wantSubjectText)
			}
			if effect.PowerDelta != test.wantPower || effect.ToughnessDelta != test.wantToughness {
				t.Fatalf("PT = %+v / %+v, want %+v / %+v", effect.PowerDelta, effect.ToughnessDelta, test.wantPower, test.wantToughness)
			}
		})
	}
}

func TestCompileStaticKeywordGrantSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source             string
		wantSubject        StaticSubjectKind
		wantSubjectSubtype string
		keywords           []string
	}{
		"enchanted creature": {
			source:      "Enchanted creature has menace.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Menace"},
		},
		"equipped creature": {
			source:      "Equipped creature has flying and first strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Flying", "First strike"},
		},
		"double strike": {
			source:      "Equipped creature has double strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Double strike"},
		},
		"other creatures": {
			source:      "Other creatures you control have flying.",
			wantSubject: StaticSubjectOtherControlledCreatures,
			keywords:    []string{"Flying"},
		},
		"controlled creatures": {
			source:      "Creatures you control have haste.",
			wantSubject: StaticSubjectControlledCreatures,
			keywords:    []string{"Haste"},
		},
		"controlled artifacts": {
			source:      "Artifacts you control have indestructible.",
			wantSubject: StaticSubjectControlledArtifacts,
			keywords:    []string{"Indestructible"},
		},
		"controlled subtype": {
			source:             "Zombies you control have flying.",
			wantSubject:        StaticSubjectControlledCreatureSubtype,
			wantSubjectSubtype: "Zombies",
			keywords:           []string{"Flying"},
		},
		"other controlled subtype": {
			source:             "Other Dinosaurs you control have haste.",
			wantSubject:        StaticSubjectOtherControlledCreatureSubtype,
			wantSubjectSubtype: "Dinosaurs",
			keywords:           []string{"Haste"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectGrantKeyword {
				t.Fatalf("effects = %#v", ability.Effects)
			}
			if got := ability.Effects[0].StaticSubject; got != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", got, test.wantSubject)
			}
			if got := ability.Effects[0].StaticSubjectSubtype; got != test.wantSubjectSubtype {
				t.Fatalf("static subject subtype = %q, want %q", got, test.wantSubjectSubtype)
			}
			if len(ability.Keywords) != len(test.keywords) {
				t.Fatalf("keywords = %#v, want %v", ability.Keywords, test.keywords)
			}
			for i, keyword := range ability.Keywords {
				if keyword.Name != test.keywords[i] {
					t.Fatalf("keyword %d = %q, want %q", i, keyword.Name, test.keywords[i])
				}
			}
		})
	}
}

func TestCompileStaticPTBuffWithKeywordHasOneEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Creatures you control get +1/+1 and have vigilance.",
		ParseContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectModifyPT {
		t.Fatalf("effects = %#v", ability.Effects)
	}
}

func TestCompileResolvingPTBuffHasNoStaticSubject(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Target creature gets +2/+2 until end of turn.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Effects[0]
	if effect.StaticSubject != StaticSubjectNone {
		t.Fatalf("static subject = %v, want StaticSubjectNone", effect.StaticSubject)
	}
	if effect.StaticSubjectSpan != (Span{}) {
		t.Fatalf("static subject span = %#v, want zero span", effect.StaticSubjectSpan)
	}
}

func TestCompileSurveilEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Surveil 2.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectSurveil ||
		effects[0].Amount != (CompiledAmount{Value: 2, Known: true}) {
		t.Fatalf("effects = %#v, want surveil 2", effects)
	}
}

func TestCompileInvestigateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Investigate.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectInvestigate {
		t.Fatalf("effects = %#v, want investigate", effects)
	}
}

func TestCompileProliferateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Proliferate.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectProliferate {
		t.Fatalf("effects = %#v, want proliferate", effects)
	}
}

func TestCompileRegenerateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Regenerate target creature.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectRegenerate {
		t.Fatalf("effects = %#v, want regenerate", effects)
	}
}

func TestCompileCounterVerbAndNoun(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wantKinds []EffectKind
	}{
		"Counter target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"This spell counters target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"Put two +1/+1 counters on target creature.": {
			wantKinds: []EffectKind{EffectPut},
		},
		"Remove a counter from this permanent: Draw a card.": {
			wantKinds: []EffectKind{EffectDraw},
		},
	}
	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Effects
			if len(effects) != len(test.wantKinds) {
				t.Fatalf("effects = %#v, want kinds %v", effects, test.wantKinds)
			}
			for i, want := range test.wantKinds {
				if effects[i].Kind != want {
					t.Fatalf("effect %d = %v, want %v", i, effects[i].Kind, want)
				}
			}
		})
	}
}

func TestCompileNegatedEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Players can't gain life.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectGain || !effects[0].Negated {
		t.Fatalf("effects = %#v", effects)
	}
}

func TestCompileEntersTappedUnlessCondition(t *testing.T) {
	t.Parallel()
	source := "This land enters tapped unless you control two or more basic lands."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != EffectEnterTapped {
		t.Fatalf("effects = %#v", ability.Effects)
	}
	if len(ability.Conditions) != 1 ||
		ability.Conditions[0].Kind != ConditionUnless ||
		ability.Conditions[0].Text != "unless you control two or more basic lands" {
		t.Fatalf("conditions = %#v", ability.Conditions)
	}
	if len(ability.References) != 1 || ability.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.References)
	}
}

func TestCompileArtifactAndEnchantmentEntersTappedReference(t *testing.T) {
	t.Parallel()
	tests := []string{
		"This artifact enters tapped.",
		"This enchantment enters tapped.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Kind != AbilityReplacement {
				t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
			}
			if len(ability.References) != 1 || ability.References[0].Kind != ReferenceThisObject {
				t.Fatalf("references = %#v", ability.References)
			}
		})
	}
}

func TestCompileUnsupportedConstruct(t *testing.T) {
	t.Parallel()
	source := "Start your engines!"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	if len(diagnostics) != 1 ||
		diagnostics[0].Severity != SeverityWarning ||
		diagnostics[0].Span != compilation.Abilities[0].Span {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestCompileScryfallCacheHasNoSilentAbilities(t *testing.T) {
	t.Parallel()
	cache := filepath.Join("..", "..", ".cardwork", "deck", "cache", "scryfall")
	paths, err := filepath.Glob(filepath.Join(cache, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("local Scryfall cache is not present")
	}

	var texts int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var card cachedParserCard
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		check := func(name, typeLine, source string) {
			t.Helper()
			if source == "" {
				return
			}
			texts++
			context := ParseContext{
				CardName:         name,
				InstantOrSorcery: typeLine == "Instant" || typeLine == "Sorcery",
				Planeswalker:     typeLine == "Planeswalker" || typeLine == "Legendary Planeswalker",
			}
			compilation, diagnostics := Compile(source, context)
			for _, diagnostic := range diagnostics {
				if diagnostic.Severity == SeverityError {
					t.Fatalf("%s: compiler error = %#v", name, diagnostic)
				}
			}
			for _, ability := range compilation.Abilities {
				if ability.Kind == AbilityReminder {
					continue
				}
				meaningful := len(ability.Effects) > 0 ||
					len(ability.Keywords) > 0 ||
					len(ability.Modes) > 0
				if meaningful || hasDiagnosticForSpan(diagnostics, ability.Span) {
					continue
				}
				t.Fatalf("%s: silently uncompiled ability %q", name, ability.Text)
			}
		}
		check(card.Name, card.TypeLine, card.OracleText)
		for _, face := range card.CardFaces {
			check(face.Name, face.TypeLine, face.OracleText)
		}
	}
	if texts != 59 {
		t.Fatalf("checked %d non-empty Oracle texts, want 59", texts)
	}
}

func hasDiagnosticForSpan(diagnostics []Diagnostic, span Span) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Span == span {
			return true
		}
	}
	return false
}
