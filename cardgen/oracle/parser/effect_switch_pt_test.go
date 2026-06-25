package parser

import "testing"

func TestParseSwitchPowerToughnessSourceForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, source string
		ctx          Context
	}{
		{
			name:   "Aeromoeba",
			source: "Discard a card: Switch this creature's power and toughness until end of turn.",
			ctx:    Context{},
		},
		{
			name:   "Crag Puca",
			source: "{U/R}: Switch this creature's power and toughness until end of turn.",
			ctx:    Context{},
		},
		{
			name:   "Flatman",
			source: "{2}{G}: Switch Flatman's power and toughness until end of turn.",
			ctx:    Context{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := test.ctx
			ctx.CardName = test.name
			document, diagnostics := Parse(test.source, ctx)
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", test.source, diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
				t.Fatalf("Parse(%q) shape = %#v", test.source, document.Abilities)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("Parse(%q) effects = %#v", test.source, effects)
			}
			effect := effects[0]
			if effect.Kind != EffectSwitchPT {
				t.Fatalf("kind = %v, want EffectSwitchPT", effect.Kind)
			}
			if !effect.SwitchPTSource {
				t.Fatal("SwitchPTSource = false, want true")
			}
			if effect.Duration != EffectDurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", effect.Duration)
			}
		})
	}
}

// TestParseSwitchPowerToughnessFailsClosed confirms the recognizer leaves the
// target, group, and missing-duration forms unrecognized so they fail closed.
func TestParseSwitchPowerToughnessFailsClosed(t *testing.T) {
	t.Parallel()
	sources := []string{
		"Switch target creature's power and toughness until end of turn.",
		"Switch each creature's power and toughness until end of turn.",
		"Switch this creature's power and toughness.",
	}
	for _, source := range sources {
		document, _ := Parse(source, Context{CardName: "Test", InstantOrSorcery: true})
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectSwitchPT {
						t.Fatalf("Parse(%q) unexpectedly recognized EffectSwitchPT", source)
					}
				}
			}
		}
	}
}
