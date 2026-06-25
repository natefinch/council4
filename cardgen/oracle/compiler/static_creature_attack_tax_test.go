package compiler

import "testing"

func TestCompileStaticCreatureAttackTaxDeclaration(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source               string
		card                 string
		amount               StaticCreatureAttackTaxAmountKind
		fixedGeneric         int
		includePlaneswalkers bool
	}{
		"enchantment scaled (Sphere of Safety)": {
			source:               "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.",
			card:                 "Sphere of Safety",
			amount:               StaticCreatureAttackTaxEnchantments,
			includePlaneswalkers: true,
		},
		"fixed planeswalker-inclusive (Baird)": {
			source:               "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.",
			card:                 "Baird, Steward of Argive",
			amount:               StaticCreatureAttackTaxFixed,
			fixedGeneric:         1,
			includePlaneswalkers: true,
		},
		"domain player-only (Collective Restraint)": {
			source: "Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control.",
			card:   "Collective Restraint",
			amount: StaticCreatureAttackTaxDomain,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.source, pipelineContext{CardName: tc.card})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 {
				t.Fatalf("abilities = %#v, want one", compilation.Abilities)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil || len(ability.Static.Declarations) != 1 {
				t.Fatalf("static = %#v, want one declaration", ability.Static)
			}
			declaration := ability.Static.Declarations[0]
			if declaration.Kind != StaticDeclarationCreatureAttackTax {
				t.Fatalf("kind = %v, want creature attack tax", declaration.Kind)
			}
			if declaration.CreatureAttackTax == nil {
				t.Fatalf("payload = %#v, want creature attack tax payload", declaration)
			}
			tax := declaration.CreatureAttackTax
			if tax.Amount != tc.amount {
				t.Fatalf("amount = %v, want %v", tax.Amount, tc.amount)
			}
			if tax.FixedGeneric != tc.fixedGeneric {
				t.Fatalf("fixed generic = %d, want %d", tax.FixedGeneric, tc.fixedGeneric)
			}
			if tax.IncludePlaneswalkers != tc.includePlaneswalkers {
				t.Fatalf("include planeswalkers = %v, want %v", tax.IncludePlaneswalkers, tc.includePlaneswalkers)
			}
		})
	}
}

func TestCompileStaticCreatureAttackTaxAbsent(t *testing.T) {
	t.Parallel()
	// A fixed-generic Propaganda tax must not produce the creature attack tax.
	source := "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."
	compilation, _ := compileSource(source, pipelineContext{CardName: "Propaganda"})
	for _, ability := range compilation.Abilities {
		if ability.Static == nil {
			continue
		}
		for _, declaration := range ability.Static.Declarations {
			if declaration.Kind == StaticDeclarationCreatureAttackTax {
				t.Fatal("Propaganda matched the creature attack tax")
			}
		}
	}
}
