package parser

import "testing"

// setBasePTEffect parses a single-sentence source and returns its lone resolving
// effect, asserting a clean single-ability single-sentence single-effect shape.
func setBasePTEffect(t *testing.T, name, source string, ctx Context) EffectSyntax {
	t.Helper()
	ctx.CardName = name
	document, diagnostics := Parse(source, ctx)
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

func TestParseSetBasePowerToughnessGroupVariableX(t *testing.T) {
	t.Parallel()
	effect := setBasePTEffect(t, "Mirror Entity",
		"{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.",
		Context{})
	if effect.Kind != EffectSetBasePT {
		t.Fatalf("kind = %v, want EffectSetBasePT", effect.Kind)
	}
	if !effect.SetBasePTVariableX {
		t.Fatal("SetBasePTVariableX = false, want true")
	}
	if !effect.SetBasePTEveryCreatureType {
		t.Fatal("SetBasePTEveryCreatureType = false, want true")
	}
	if effect.StaticSubject.Kind == EffectStaticSubjectNone {
		t.Fatalf("StaticSubject = %v, want a controlled-creature group", effect.StaticSubject)
	}
}

func TestParseSetBasePowerToughnessForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, source       string
		ctx                Context
		wantVariableX      bool
		wantEveryType      bool
		wantSource         bool
		wantPower, wantTgh int
	}{
		{
			name:   "Square Up",
			source: "Target creature has base power and toughness 4/4 until end of turn.",
			ctx:    Context{InstantOrSorcery: true}, wantPower: 4, wantTgh: 4,
		},
		{
			name:   "Biomass Mutation",
			source: "Creatures you control have base power and toughness X/X until end of turn.",
			ctx:    Context{InstantOrSorcery: true}, wantVariableX: true,
		},
		{
			name:   "Marsh Flitter",
			source: "This creature has base power and toughness 3/3 until end of turn.",
			ctx:    Context{}, wantSource: true, wantPower: 3, wantTgh: 3,
		},
	}
	for _, test := range tests {
		effect := setBasePTEffect(t, test.name, test.source, test.ctx)
		if effect.Kind != EffectSetBasePT {
			t.Errorf("%s: kind = %v, want EffectSetBasePT", test.name, effect.Kind)
			continue
		}
		if effect.SetBasePTVariableX != test.wantVariableX {
			t.Errorf("%s: SetBasePTVariableX = %v, want %v", test.name, effect.SetBasePTVariableX, test.wantVariableX)
		}
		if effect.SetBasePTEveryCreatureType != test.wantEveryType {
			t.Errorf("%s: SetBasePTEveryCreatureType = %v, want %v", test.name, effect.SetBasePTEveryCreatureType, test.wantEveryType)
		}
		if effect.SetBasePTSource != test.wantSource {
			t.Errorf("%s: SetBasePTSource = %v, want %v", test.name, effect.SetBasePTSource, test.wantSource)
		}
		if !test.wantVariableX {
			if effect.SetBasePower != test.wantPower || effect.SetBaseToughness != test.wantTgh {
				t.Errorf("%s: base P/T = %d/%d, want %d/%d", test.name, effect.SetBasePower, effect.SetBaseToughness, test.wantPower, test.wantTgh)
			}
		}
	}
}

// TestParseSetBasePowerToughnessFailsClosed confirms shapes the executable
// backend cannot represent do not produce an EffectSetBasePT effect, so those
// cards stay unsupported rather than lowering an approximate result.
func TestParseSetBasePowerToughnessFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		// A keyword rider on the set is unrepresentable here.
		{"Trample Set", "Target creature has base power and toughness 4/4 and gains trample until end of turn."},
		// "where X is ..." dynamic counts are out of scope.
		{"Counted Set", "Creatures you control have base power and toughness X/X until end of turn, where X is the number of Zombies you control."},
		// Missing the until-end-of-turn duration is the permanent static
		// declaration form, parsed elsewhere.
		{"Godhead", "Other creatures have base power and toughness 1/1."},
		// A subtype change rider is unrepresentable here.
		{"Subtype Set", "Target creature has base power and toughness 0/1 and becomes a Wall until end of turn."},
	}
	for _, test := range tests {
		document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true, CardName: test.name})
		if len(diagnostics) != 0 {
			continue
		}
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectSetBasePT {
						t.Errorf("%s: produced EffectSetBasePT, want fail closed", test.name)
					}
				}
			}
		}
	}
}
