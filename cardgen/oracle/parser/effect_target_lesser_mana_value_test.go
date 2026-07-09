package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// firstReturnTarget parses a single graveyard-return sentence and returns its
// lone target's selection, failing the test on any diagnostic or unexpected
// shape.
func firstReturnTarget(t *testing.T, source string) SelectionSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("Parse(%q) targets = %#v, want one", source, targets)
	}
	return targets[0].Selection
}

// TestParseLesserManaValueGraveyardReturnTargetSetsEventBound proves the bare
// "with lesser mana value" clause on a graveyard-return card target records the
// event-relative ManaValueLessThanEventPermanent flag (Orah, Skyclave
// Hierophant) while keeping the graveyard zone and subtype filter, and does not
// spuriously set a fixed mana-value comparison.
func TestParseLesserManaValueGraveyardReturnTargetSetsEventBound(t *testing.T) {
	t.Parallel()
	sel := firstReturnTarget(t, "Return target Cleric card with lesser mana value from your graveyard to the battlefield.")
	if !sel.ManaValueLessThanEventPermanent {
		t.Fatal("bare \"with lesser mana value\" must set ManaValueLessThanEventPermanent")
	}
	if sel.MatchManaValue {
		t.Fatal("the event-relative bound must not set a fixed mana-value comparison")
	}
	if sel.Zone != zone.Graveyard {
		t.Fatalf("target zone = %v, want Graveyard", sel.Zone)
	}
	if len(sel.SubtypesAny) != 1 || sel.SubtypesAny[0] != "Cleric" {
		t.Fatalf("subtypes = %v, want [Cleric]", sel.SubtypesAny)
	}
}

// TestParseLesserManaValueBoundExcludesNonEventForms proves the parser fails
// closed on the near-miss mana-value wordings that are not the event-relative
// strict bound: "equal or lesser mana value" (a ≤ bound) and the explicit
// "lesser mana value than <object>" comparison used by seek/cascade effects must
// not set ManaValueLessThanEventPermanent.
func TestParseLesserManaValueBoundExcludesNonEventForms(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Return target Cleric card with equal or lesser mana value from your graveyard to the battlefield.",
		"Return target Cleric card with lesser mana value than the exiled card from your graveyard to the battlefield.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, target := range sentence.Targets {
					if target.Selection.ManaValueLessThanEventPermanent {
						t.Fatalf("Parse(%q) set ManaValueLessThanEventPermanent, want unrecognized (fail closed)", source)
					}
				}
			}
		}
	}
}
