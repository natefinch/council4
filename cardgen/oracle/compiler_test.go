package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
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
