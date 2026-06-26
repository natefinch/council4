package parser

import "testing"

func TestParseStaticManaProductionMultiplierDeclarationMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source string
		card   string
		factor int
	}{
		"doubler (Mana Reflection)": {
			source: "If you tap a permanent for mana, it produces twice as much of that mana instead.",
			card:   "Mana Reflection",
			factor: 2,
		},
		"tripler (Nyxbloom Ancient)": {
			source: "If you tap a permanent for mana, it produces three times as much of that mana instead.",
			card:   "Nyxbloom Ancient",
			factor: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{CardName: tc.card})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationManaProductionMultiplier {
				t.Fatalf("kind = %v, want mana production multiplier", declaration.Kind)
			}
			if declaration.ManaMultiplier != tc.factor {
				t.Fatalf("factor = %d, want %d", declaration.ManaMultiplier, tc.factor)
			}
		})
	}
}

func TestParseStaticManaProductionMultiplierDeclarationFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// "once" would scale by one, which carries no replacement and is rejected.
		"factor below two": "If you tap a permanent for mana, it produces as much of that mana instead.",
		// A different tapped object than "a permanent" is a narrower trigger.
		"land only": "If you tap a land for mana, it produces twice as much of that mana instead.",
		// Tapping "a creature" is a different, narrower source set.
		"creature only": "If you tap a creature for mana, it produces twice as much of that mana instead.",
		// Dropping the "instead" replacement keyword changes the meaning.
		"missing instead": "If you tap a permanent for mana, it produces twice as much of that mana.",
		// "an opponent" taps a different player than the controller.
		"opponent taps": "If an opponent taps a permanent for mana, it produces twice as much of that mana instead.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{CardName: "Tester"})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.Kind == StaticDeclarationManaProductionMultiplier {
						t.Fatalf("source %q matched the mana production multiplier", source)
					}
				}
			}
		})
	}
}
