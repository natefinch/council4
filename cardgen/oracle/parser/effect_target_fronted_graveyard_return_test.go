package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// firstReturnSentence parses a single return sentence and returns its lone
// effect and target, failing the test on any diagnostic or unexpected shape.
func firstReturnSentence(t *testing.T, source string) (EffectSyntax, TargetSyntax) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	sentence := document.Abilities[0].Sentences[0]
	if len(sentence.Effects) != 1 || len(sentence.Targets) != 1 {
		t.Fatalf("Parse(%q) effects/targets = %d/%d, want one each",
			source, len(sentence.Effects), len(sentence.Targets))
	}
	return sentence.Effects[0], sentence.Targets[0]
}

// TestParseFrontedGraveyardReturnSetsGraveyardSourceAndEventBound proves the
// fronted-destination graveyard return "Return to your hand target artifact card
// in your graveyard with lesser mana value." (Scrap Trawler) resolves the
// graveyard source from its inline "in your graveyard" phrase rather than a
// "from" preposition, reconstructs byte-exact, and records the event-relative
// ManaValueLessThanEventPermanent bound alongside the graveyard zone, "your"
// owner, and artifact filter without setting a fixed mana-value comparison.
func TestParseFrontedGraveyardReturnSetsGraveyardSourceAndEventBound(t *testing.T) {
	t.Parallel()
	effect, target := firstReturnSentence(t,
		"Return to your hand target artifact card in your graveyard with lesser mana value.")
	if effect.Kind != EffectReturn {
		t.Fatalf("effect kind = %v, want EffectReturn", effect.Kind)
	}
	if effect.FromZone != zone.Graveyard || effect.ToZone != zone.Hand {
		t.Fatalf("effect zones = from %v to %v, want graveyard to hand", effect.FromZone, effect.ToZone)
	}
	if !effect.Exact {
		t.Fatal("fronted graveyard return must reconstruct byte-exact (Exact)")
	}
	sel := target.Selection
	if sel.Zone != zone.Graveyard {
		t.Fatalf("target zone = %v, want Graveyard", sel.Zone)
	}
	if !sel.ManaValueLessThanEventPermanent {
		t.Fatal("trailing \"with lesser mana value\" must set ManaValueLessThanEventPermanent")
	}
	if sel.MatchManaValue {
		t.Fatal("the event-relative bound must not set a fixed mana-value comparison")
	}
	if sel.Controller != SelectionControllerYou {
		t.Fatalf("controller = %v, want SelectionControllerYou", sel.Controller)
	}
	if len(sel.RequiredTypesAny) != 1 || sel.RequiredTypesAny[0] != CardTypeArtifact {
		t.Fatalf("required types = %v, want [CardTypeArtifact]", sel.RequiredTypesAny)
	}
}

// TestParseFrontedGraveyardReturnWithoutManaValueBound proves the fronted
// graveyard return still resolves its graveyard source and byte-exact
// reconstruction when the noun clause carries no "with lesser mana value"
// qualifier, and leaves the event-relative bound unset.
func TestParseFrontedGraveyardReturnWithoutManaValueBound(t *testing.T) {
	t.Parallel()
	effect, target := firstReturnSentence(t,
		"Return to your hand target artifact card in your graveyard.")
	if effect.FromZone != zone.Graveyard || effect.ToZone != zone.Hand || !effect.Exact {
		t.Fatalf("effect = from %v to %v exact %v, want graveyard->hand exact",
			effect.FromZone, effect.ToZone, effect.Exact)
	}
	if target.Selection.Zone != zone.Graveyard {
		t.Fatalf("target zone = %v, want Graveyard", target.Selection.Zone)
	}
	if target.Selection.ManaValueLessThanEventPermanent {
		t.Fatal("no \"lesser mana value\" clause must leave the event bound unset")
	}
}

// TestParseFrontedGraveyardReturnEqualOrLesserFailsClosed proves the ≤ near-miss
// wording "with equal or lesser mana value" does not absorb into the strict
// event-relative bound in the fronted form, mirroring the target-trailing form's
// fail-closed handling.
func TestParseFrontedGraveyardReturnEqualOrLesserFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Return to your hand target artifact card in your graveyard with equal or lesser mana value.",
		Context{InstantOrSorcery: true})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, target := range sentence.Targets {
				if target.Selection.ManaValueLessThanEventPermanent {
					t.Fatal("\"equal or lesser mana value\" must not set the strict event bound")
				}
			}
		}
	}
}
