package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
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
