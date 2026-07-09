package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// deathsPresenceText is the Death's Presence ability whose "where X is the power
// of the creature that died" amount reads the dying creature's power through
// last-known information.
const deathsPresenceText = "Whenever a creature you control dies, put X +1/+1 counters on target creature you control, where X is the power of the creature that died."

// singleDiedPutEffect returns the sole EffectPut in a parsed document, failing
// the test if the document does not contain exactly one.
func singleDiedPutEffect(t *testing.T, document Document) EffectSyntax {
	t.Helper()
	var puts []EffectSyntax
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectPut {
					puts = append(puts, effect)
				}
			}
		}
	}
	if len(puts) != 1 {
		t.Fatalf("put effects = %d, want 1", len(puts))
	}
	return puts[0]
}

// TestParseEmitsDiedCreatureReference proves "the power of the creature that
// died" reports "the creature that died" as a dedicated ReferenceDiedCreature.
// The dedicated kind keeps it distinct from a demonstrative "that creature", so
// the compiler binds it to the dying permanent rather than a target antecedent
// (Death's Presence).
func TestParseEmitsDiedCreatureReference(t *testing.T) {
	t.Parallel()
	refs := atomsFor(t, deathsPresenceText, "Death's Presence").References()
	var died []Reference
	for _, ref := range refs {
		if ref.Kind == ReferenceDiedCreature {
			died = append(died, ref)
		}
	}
	if len(died) != 1 {
		t.Fatalf("died-creature references = %+v; want exactly one", refs)
	}
	if got := shared.SliceSpan(deathsPresenceText, died[0].Span); got != "the creature that died" {
		t.Errorf("died-creature span = %q; want %q", got, "the creature that died")
	}
}

// TestParseComparisonCreatureThatDiedNotDiedReference proves the comparison
// wording "lesser mana value than the creature that died" reports no
// died-creature reference. The recognizer is scoped to the "the
// power/toughness/mana value of the creature that died" characteristic form, so
// unrelated "the creature that died" mentions such as Death's Oasis stay
// unaffected and fail-closed.
func TestParseComparisonCreatureThatDiedNotDiedReference(t *testing.T) {
	t.Parallel()
	source := "Whenever a nontoken creature you control dies, mill two cards. Then return a creature card with lesser mana value than the creature that died from your graveyard to your hand."
	document, _ := Parse(source, Context{CardName: "Death's Oasis"})
	for _, ability := range document.Abilities {
		for _, ref := range ability.Atoms.References() {
			if ref.Kind == ReferenceDiedCreature {
				t.Fatalf("unexpected died-creature reference: %+v", ref)
			}
		}
	}
}

// TestParseDiedCreaturePowerAmount proves "where X is the power of the creature
// that died" types an exact referenced-object power amount whose ReferenceSpan
// covers "the creature that died" — the shape the compiler binds to the event
// permanent and the lowerer reads through last-known information (Death's
// Presence).
func TestParseDiedCreaturePowerAmount(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(deathsPresenceText, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	effect := singleDiedPutEffect(t, document)
	if !effect.Exact {
		t.Fatalf("put effect not exact: %#v", effect)
	}
	if effect.Amount.DynamicKind != EffectDynamicAmountSourcePower {
		t.Fatalf("amount kind = %q, want %q", effect.Amount.DynamicKind, EffectDynamicAmountSourcePower)
	}
	if got := shared.SliceSpan(deathsPresenceText, effect.Amount.ReferenceSpan); got != "the creature that died" {
		t.Errorf("amount reference span = %q, want %q", got, "the creature that died")
	}
}
