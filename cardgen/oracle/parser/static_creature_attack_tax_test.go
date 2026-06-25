package parser

import "testing"

const (
	sphereOfSafetyText     = "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control."
	bairdStewardText       = "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures."
	collectiveRestraintTax = "Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control."
)

func TestParseStaticCreatureAttackTaxDeclarationMeaning(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		source               string
		card                 string
		amount               StaticAttackTaxAmountKind
		generic              int
		includePlaneswalkers bool
	}{
		"enchantment scaled (Sphere of Safety)": {
			source:               sphereOfSafetyText,
			card:                 "Sphere of Safety",
			amount:               StaticAttackTaxAmountEnchantments,
			includePlaneswalkers: true,
		},
		"fixed planeswalker-inclusive (Baird)": {
			source:               bairdStewardText,
			card:                 "Baird, Steward of Argive",
			amount:               StaticAttackTaxAmountFixed,
			generic:              1,
			includePlaneswalkers: true,
		},
		"domain player-only (Collective Restraint)": {
			source: collectiveRestraintTax,
			card:   "Collective Restraint",
			amount: StaticAttackTaxAmountDomain,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{CardName: tc.card})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationCreatureAttackTax {
				t.Fatalf("kind = %v, want creature attack tax", declaration.Kind)
			}
			if declaration.AttackTaxAmountKind != tc.amount {
				t.Fatalf("amount kind = %v, want %v", declaration.AttackTaxAmountKind, tc.amount)
			}
			if declaration.AttackTaxGeneric != tc.generic {
				t.Fatalf("generic = %d, want %d", declaration.AttackTaxGeneric, tc.generic)
			}
			if declaration.AttackTaxIncludesPlaneswalkers != tc.includePlaneswalkers {
				t.Fatalf("includes planeswalkers = %v, want %v", declaration.AttackTaxIncludesPlaneswalkers, tc.includePlaneswalkers)
			}
		})
	}
}

func TestParseStaticCreatureAttackTaxDeclarationFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// A different counted permanent type must not match the enchantment form.
		"artifacts": "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of artifacts you control.",
		// A {X} cost without the explaining clause has no per-attacker amount.
		"scaled without clause": "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures.",
		// "opponents control" is a different counted set than "you control".
		"opponent count": "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments your opponents control.",
		// A domain clause counting something other than basic land types differs.
		"domain non-basic": "Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of lands you control.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{CardName: "Tester"})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.Kind == StaticDeclarationCreatureAttackTax {
						t.Fatalf("source %q matched the creature attack tax", source)
					}
				}
			}
		})
	}
}
