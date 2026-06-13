package parser

import (
	"testing"
)

// parseStaticDeclarationSyntax parses a single static-declaration ability and
// returns the typed declarations the parser emitted. It fails the test when the
// source produced anything other than exactly one ability so meaning tests
// assert on fully typed syntax rather than source text.
func parseStaticDeclarationSyntax(t *testing.T, source string, context Context) []StaticDeclarationSyntax {
	t.Helper()
	document, diagnostics := Parse(source, context)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want exactly one", document.Abilities)
	}
	return document.Abilities[0].StaticDeclarations()
}

func TestParseStaticPowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "This creature gets +1/+2.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousPowerToughness {
		t.Fatalf("kind = %v, want power/toughness", declaration.Kind)
	}
	if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
		t.Fatalf("subject = %#v, want source creature", declaration.Subject)
	}
	if declaration.PowerDelta.Value != 1 || declaration.PowerDelta.Negative ||
		declaration.ToughnessDelta.Value != 2 || declaration.ToughnessDelta.Negative ||
		declaration.Dynamic {
		t.Fatalf("declaration = %#v, want +1/+2 static", declaration)
	}
}

func TestParseStaticGroupPowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "Creatures you control get +1/+1.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationContinuousPowerToughness ||
		declaration.Subject.Kind != StaticDeclarationSubjectGroup {
		t.Fatalf("declaration = %#v, want group power/toughness", declaration)
	}
}

func TestParseStaticKeywordGrantDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"As long as you have 7 or more life, this creature has flying.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationKeywordGrant ||
		declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature ||
		!declaration.HasCondition ||
		len(declaration.KeywordSpans) != 1 {
		t.Fatalf("declaration = %#v, want conditional keyword grant", declaration)
	}
}

func TestParseStaticPowerToughnessAndKeywordComposition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Creatures you control get +1/+1 and have trample.",
		Context{},
	)
	if len(declarations) != 2 {
		t.Fatalf("declarations = %#v, want two", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
		declarations[1].Kind != StaticDeclarationKeywordGrant {
		t.Fatalf("declarations = %#v, want PT then keyword", declarations)
	}
	if declarations[0].Dynamic {
		t.Fatalf("composed PT declaration must not be dynamic: %#v", declarations[0])
	}
}

func TestParseStaticMultipleKeywordListComposition(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(
		t,
		"Creatures you control have flying, vigilance, and trample.",
		Context{},
	)
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	if declarations[0].Kind != StaticDeclarationKeywordGrant ||
		len(declarations[0].KeywordSpans) != 3 {
		t.Fatalf("declaration = %#v, want three granted keywords", declarations[0])
	}
}

func TestParseStaticMixedSourceDeclarationComposition(t *testing.T) {
	t.Parallel()
	source := "Delirium — As long as there are four or more card types among cards in your graveyard, " +
		"Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able."
	declarations := parseStaticDeclarationSyntax(t, source, Context{CardName: "Dragon's Rage Channeler"})
	if len(declarations) != 3 {
		t.Fatalf("declarations = %#v, want three", declarations)
	}
	if declarations[0].Kind != StaticDeclarationContinuousPowerToughness ||
		declarations[1].Kind != StaticDeclarationKeywordGrant ||
		declarations[2].Kind != StaticDeclarationRule {
		t.Fatalf("declarations = %#v, want PT, keyword, rule", declarations)
	}
	if declarations[2].Rule.Operation.Kind != StaticRuleOperationAttack {
		t.Fatalf("rule = %#v, want attack requirement", declarations[2].Rule)
	}
	for i, declaration := range declarations {
		if declaration.Subject.Kind != StaticDeclarationSubjectSourceNamed || !declaration.HasCondition {
			t.Fatalf("declaration %d = %#v, want conditional self-name subject", i, declaration)
		}
	}
}

func TestParseStaticRuleDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source    string
		subject   StaticDeclarationSubjectKind
		operation StaticRuleOperationKind
		voice     StaticRuleVoice
	}{
		"cannot block": {
			source:    "This creature can't block.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoiceActive,
		},
		"cannot be blocked": {
			source:    "This creature can't be blocked.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationBlock,
			voice:     StaticRuleVoicePassive,
		},
		"must attack": {
			source:    "This creature attacks each combat if able.",
			subject:   StaticDeclarationSubjectSourceCreature,
			operation: StaticRuleOperationAttack,
			voice:     StaticRuleVoiceActive,
		},
		"cannot be countered": {
			source:    "This spell can't be countered.",
			subject:   StaticDeclarationSubjectSourceSpell,
			operation: StaticRuleOperationCounter,
			voice:     StaticRuleVoicePassive,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationRule {
				t.Fatalf("declarations = %#v, want one rule", declarations)
			}
			rule := declarations[0].Rule
			if rule.Operation.Kind != test.operation || rule.Operation.Voice != test.voice {
				t.Fatalf("rule = %#v, want operation %d voice %d", rule, test.operation, test.voice)
			}
			if declarations[0].Subject.Kind != test.subject {
				t.Fatalf("subject = %#v, want %d", declarations[0].Subject, test.subject)
			}
		})
	}
}

func TestParseStaticCostModifierDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source      string
		modifier    StaticDeclarationCostModifierKind
		reduction   int
		replacement string
	}{
		"ability reduction": {
			source:    "Cycling abilities you activate cost up to {2} less to activate.",
			modifier:  StaticDeclarationCostModifierAbilityReduction,
			reduction: 2,
		},
		"replace cost": {
			source:      "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.",
			modifier:    StaticDeclarationCostModifierReplaceCost,
			replacement: "",
		},
		"replace first cost": {
			source:      "You may pay {1} rather than pay the cycling cost of the first card you cycle each turn.",
			modifier:    StaticDeclarationCostModifierReplaceFirstCost,
			replacement: "{1}",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationCostModifier {
				t.Fatalf("declarations = %#v, want one cost modifier", declarations)
			}
			declaration := declarations[0]
			if declaration.CostModifier != test.modifier ||
				declaration.CostReductionAmount != test.reduction ||
				declaration.CostReplacement != test.replacement {
				t.Fatalf("declaration = %#v, want modifier %d", declaration, test.modifier)
			}
		})
	}
}

func TestParseStaticCardAbilityGrantDeclarationMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		filter StaticDeclarationCardFilterKind
	}{
		"land cards": {
			source: "Each land card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterLand,
		},
		"creature cards": {
			source: "Each creature card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterCreature,
		},
		"historic cards": {
			source: "Each historic card in your hand has cycling {2}.",
			filter: StaticDeclarationCardFilterHistoric,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 || declarations[0].Kind != StaticDeclarationCardAbilityGrant {
				t.Fatalf("declarations = %#v, want one card-ability grant", declarations)
			}
			if declarations[0].Subject.Kind != StaticDeclarationSubjectControllerHand ||
				declarations[0].Subject.CardFilter != test.filter {
				t.Fatalf("subject = %#v, want hand filter %d", declarations[0].Subject, test.filter)
			}
		})
	}
}

func TestParseStaticDeclarationsFailClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"unknown verb":             "This creature flickers +1/+1.",
		"dangling connector":       "This creature gets +1/+1 and.",
		"attack missing qualifier": "This creature attacks each combat.",
		"unsupported keyword slot": "This creature has +1/+1.",
		"group rule unsupported":   "Creatures you control can't block.",
		"trailing junk":            "This creature gets +1/+1 wobble.",
		"comma without and":        "This creature gets +1/+1, has flying.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", document.Abilities)
			}
			if declarations := document.Abilities[0].StaticDeclarations(); len(declarations) != 0 {
				t.Fatalf("declarations = %#v, want none (fail closed)", declarations)
			}
		})
	}
}
