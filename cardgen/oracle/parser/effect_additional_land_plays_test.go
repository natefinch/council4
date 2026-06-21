package parser

import (
	"testing"
)

// additionalLandPlaysEffect parses a single one-shot additional-land sentence and
// returns the typed effect, asserting the parser produced exactly one effect.
func additionalLandPlaysEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v, want one", source, effects)
	}
	return effects[0]
}

func TestParseAdditionalLandPlaysOneShot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		count  int
	}{
		{"Play an additional land this turn.", 1},
		{"You may play an additional land this turn.", 1},
		{"You may play two additional lands this turn.", 2},
		{"You may play up to three additional lands this turn.", 3},
	}
	for _, tc := range cases {
		effect := additionalLandPlaysEffect(t, tc.source)
		if effect.Kind != EffectAdditionalLandPlays {
			t.Errorf("Parse(%q) kind = %v, want EffectAdditionalLandPlays", tc.source, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) Exact = false, want true", tc.source)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("Parse(%q) duration = %v, want this turn", tc.source, effect.Duration)
		}
		if !effect.Amount.Known || effect.Amount.Value != tc.count {
			t.Errorf("Parse(%q) amount = %#v, want %d", tc.source, effect.Amount, tc.count)
		}
	}
}

func TestParseAdditionalLandPlaysOneShotFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"You may play an additional land.",
		"You may play two additional land this turn.",
		"You may play an additional lands this turn.",
		"Play an additional creature this turn.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		for _, effect := range effects {
			if effect.Kind == EffectAdditionalLandPlays && effect.Exact {
				t.Errorf("Parse(%q) matched additional-land effect, want fail-closed", source)
			}
		}
	}
}

func TestParseStaticAdditionalLandPlaysDeclarationMeaning(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		count  int
	}{
		{"You may play an additional land on each of your turns.", 1},
		{"You may play two additional lands on each of your turns.", 2},
	}
	for _, tc := range cases {
		declarations := parseStaticDeclarationSyntax(t, tc.source, Context{})
		if len(declarations) != 1 {
			t.Fatalf("Parse(%q) declarations = %#v, want one", tc.source, declarations)
		}
		declaration := declarations[0]
		if declaration.Kind != StaticDeclarationPlayerRule {
			t.Errorf("Parse(%q) kind = %v, want player rule", tc.source, declaration.Kind)
		}
		if declaration.Subject.Kind != StaticDeclarationSubjectController {
			t.Errorf("Parse(%q) subject = %#v, want controller", tc.source, declaration.Subject)
		}
		if declaration.PlayerRule != StaticDeclarationPlayerRuleAdditionalLandPlays {
			t.Errorf("Parse(%q) player rule = %v, want additional land plays", tc.source, declaration.PlayerRule)
		}
		if declaration.AdditionalLandPlays != tc.count {
			t.Errorf("Parse(%q) count = %d, want %d", tc.source, declaration.AdditionalLandPlays, tc.count)
		}
	}
}

func TestParseStaticEachPlayerAdditionalLandPlaysDeclarationMeaning(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		count  int
	}{
		{"Each player may play an additional land on each of their turns.", 1},
		{"Each player may play two additional lands on each of their turns.", 2},
	}
	for _, tc := range cases {
		declarations := parseStaticDeclarationSyntax(t, tc.source, Context{})
		if len(declarations) != 1 {
			t.Fatalf("Parse(%q) declarations = %#v, want one", tc.source, declarations)
		}
		declaration := declarations[0]
		if declaration.Kind != StaticDeclarationPlayerRule {
			t.Errorf("Parse(%q) kind = %v, want player rule", tc.source, declaration.Kind)
		}
		if declaration.Subject.Kind != StaticDeclarationSubjectEachPlayer {
			t.Errorf("Parse(%q) subject = %#v, want each player", tc.source, declaration.Subject)
		}
		if declaration.PlayerRule != StaticDeclarationPlayerRuleAdditionalLandPlays {
			t.Errorf("Parse(%q) player rule = %v, want additional land plays", tc.source, declaration.PlayerRule)
		}
		if declaration.AdditionalLandPlays != tc.count {
			t.Errorf("Parse(%q) count = %d, want %d", tc.source, declaration.AdditionalLandPlays, tc.count)
		}
	}
}

func TestParseStaticAdditionalLandPlaysDeclarationFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"You may play an additional land on each turn.",
		"You may play two additional land on each of your turns.",
		"You may play an additional land on each of their turns.",
	} {
		document, _ := Parse(source, Context{})
		for _, ability := range document.Abilities {
			for _, declaration := range ability.StaticDeclarations {
				if declaration.PlayerRule == StaticDeclarationPlayerRuleAdditionalLandPlays {
					t.Errorf("Parse(%q) matched additional-land static rule, want fail-closed", source)
				}
			}
		}
	}
}
