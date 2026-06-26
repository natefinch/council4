package compiler

import "testing"

func TestCompileStaticManaProductionMultiplierDeclaration(t *testing.T) {
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
			source: "Trample\nIf you tap a permanent for mana, it produces three times as much of that mana instead.",
			card:   "Nyxbloom Ancient",
			factor: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.source, pipelineContext{CardName: tc.card})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			var declaration StaticDeclaration
			found := false
			for _, ability := range compilation.Abilities {
				if ability.Static == nil {
					continue
				}
				for _, decl := range ability.Static.Declarations {
					if decl.Kind == StaticDeclarationManaProductionMultiplier {
						declaration = decl
						found = true
					}
				}
			}
			if !found {
				t.Fatalf("abilities = %#v, want a mana production multiplier declaration", compilation.Abilities)
			}
			if declaration.ManaProductionMultiplier == nil {
				t.Fatalf("payload = %#v, want mana production multiplier payload", declaration)
			}
			if declaration.ManaProductionMultiplier.Factor != tc.factor {
				t.Fatalf("factor = %d, want %d", declaration.ManaProductionMultiplier.Factor, tc.factor)
			}
		})
	}
}

func TestCompileStaticManaProductionMultiplierAbsent(t *testing.T) {
	t.Parallel()
	// A narrower "tap a land" trigger must not produce the all-permanents multiplier.
	source := "If you tap a land for mana, it produces twice as much of that mana instead."
	compilation, _ := compileSource(source, pipelineContext{CardName: "Tester"})
	for _, ability := range compilation.Abilities {
		if ability.Static == nil {
			continue
		}
		for _, declaration := range ability.Static.Declarations {
			if declaration.Kind == StaticDeclarationManaProductionMultiplier {
				t.Fatal("narrower land-only trigger matched the mana production multiplier")
			}
		}
	}
}
