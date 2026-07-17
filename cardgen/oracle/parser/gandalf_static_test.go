package parser

import (
	"testing"
)

const gandalfTriggerText = "If a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time."

func TestParseGandalfFlashPermission(t *testing.T) {
	source := "You may cast legendary spells and artifact spells as though they had flash."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("document = %#v diagnostics = %#v", document, diagnostics)
	}
	declarations := document.Abilities[0].StaticDeclarations
	if len(declarations) != 1 {
		tokens := document.Abilities[0].Tokens
		filters, ok := parseStaticSpellCharacteristicBranches(tokens[3:8])
		t.Fatalf("declarations = %#v filters = %#v ok = %v token text = %q", declarations, filters, ok, document.Abilities[0].Text)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationCastAsThoughFlash ||
		len(declaration.FlashSpellFilters) != 2 {
		t.Fatalf("declaration = %#v", declaration)
	}
	if got := declaration.FlashSpellFilters[0].Supertypes; len(got) != 1 || got[0] != SupertypeLegendary {
		t.Fatalf("legendary branch = %#v", declaration.FlashSpellFilters[0])
	}
	if got := declaration.FlashSpellFilters[1].CardTypes; len(got) != 1 || got[0] != CardTypeArtifact {
		t.Fatalf("artifact branch = %#v", declaration.FlashSpellFilters[1])
	}
}

func TestParseGandalfTriggerMultiplier(t *testing.T) {
	declarations := parseStaticDeclarationSyntax(t, gandalfTriggerText, Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationControlledTriggerMultiplier ||
		!declaration.ControlledCausePermanentEnters ||
		!declaration.ControlledCausePermanentLeaves ||
		len(declaration.ControlledCausePermanentFilters) != 2 {
		t.Fatalf("declaration = %#v", declaration)
	}
	legendary := declaration.ControlledCausePermanentFilters[0]
	if len(legendary.Supertypes) != 1 || legendary.Supertypes[0] != SupertypeLegendary ||
		len(legendary.CardTypes) != 0 {
		t.Fatalf("legendary branch = %#v", legendary)
	}
	artifact := declaration.ControlledCausePermanentFilters[1]
	if len(artifact.CardTypes) != 1 || artifact.CardTypes[0] != CardTypeArtifact ||
		len(artifact.Supertypes) != 0 || len(artifact.Subtypes) != 0 {
		t.Fatalf("artifact branch = %#v", artifact)
	}
}

func TestParseGandalfMechanicsFailClosedOnNearMisses(t *testing.T) {
	for _, source := range []string{
		"You may cast legendary cards and artifact spells as though they had flash.",
		"If a legendary permanent or an artifact entering the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
		"If a legendary permanent or an artifact entering or leaving the graveyard causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
	} {
		document, _ := Parse(source, Context{})
		for _, ability := range document.Abilities {
			for _, declaration := range ability.StaticDeclarations {
				if declaration.Kind == StaticDeclarationCastAsThoughFlash && len(declaration.FlashSpellFilters) != 0 ||
					declaration.Kind == StaticDeclarationControlledTriggerMultiplier &&
						len(declaration.ControlledCausePermanentFilters) != 0 {
					t.Fatalf("near miss %q unexpectedly parsed as Gandalf mechanic: %#v", source, declaration)
				}
			}
		}
	}
}
