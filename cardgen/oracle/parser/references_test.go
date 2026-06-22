package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func referenceKinds(references []Reference) []ReferenceKind {
	kinds := make([]ReferenceKind, len(references))
	for i := range references {
		kinds[i] = references[i].Kind
	}
	return kinds
}

// TestParseEmitsSelfNameReferenceTrailingApostrophe proves a plural-possessive
// card name whose final word ends in a bare apostrophe ("Inventors'") is still
// recognized as a self-name, so "Sacrifice <self by name>" costs resolve.
func TestParseEmitsSelfNameReferenceTrailingApostrophe(t *testing.T) {
	t.Parallel()
	source := "Sacrifice Inventors' Fair: Draw a card"
	atoms := atomsFor(t, source, "Inventors' Fair")
	refs := atoms.References()
	if kinds := referenceKinds(refs); len(kinds) != 1 || kinds[0] != ReferenceSelfName {
		t.Fatalf("references = %+v; want one self-name", refs)
	}
	if got := shared.SliceSpan(source, refs[0].Span); got != "Inventors' Fair" {
		t.Errorf("self-name span = %q; want %q", got, "Inventors' Fair")
	}
}

// TestParseEmitsSelfNameReference proves the parser recognizes the card's own
// name from Context.CardName and emits a source-spanned self-name reference.
func TestParseEmitsSelfNameReference(t *testing.T) {
	t.Parallel()
	source := "Grizzly Bears deals damage"
	atoms := atomsFor(t, source, "Grizzly Bears")
	if kinds := referenceKinds(atoms.References()); len(kinds) != 1 || kinds[0] != ReferenceSelfName {
		t.Fatalf("references = %+v; want one self-name", atoms.References())
	}
	if got := shared.SliceSpan(source, atoms.References()[0].Span); got != "Grizzly Bears" {
		t.Errorf("self-name span = %q; want %q", got, "Grizzly Bears")
	}
	// The self name is recognized from the typed card-name context, not token
	// spelling: identical tokens emit no self-name when the name differs.
	if refs := atomsFor(t, source, "Llanowar Elves").References(); len(refs) != 0 {
		t.Errorf("references with non-matching name = %+v; want none", refs)
	}
}

// TestParseEmitsSelfNameSpans proves every name occurrence, including possessive
// forms, is recorded for downstream name-blind consumers.
func TestParseEmitsSelfNameSpans(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "Nightpack Ambusher's power is doubled", "Nightpack Ambusher")
	if len(atoms.SelfNameSpans()) != 1 {
		t.Fatalf("self-name spans = %+v; want one", atoms.SelfNameSpans())
	}
	if !atoms.SelfNameStartingAt(atoms.SelfNameSpans()[0]) {
		t.Error("SelfNameStartingAt(emitted span) = false; want true")
	}
}

// TestParseEmitsThisThatPronounReferences proves explicit object references are
// emitted with their typed kinds.
func TestParseEmitsThisThatPronounReferences(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "this creature's power, that artifact, return it to their hand", "")
	want := []ReferenceKind{ReferenceThisObject, ReferenceThatObject, ReferencePronoun, ReferencePronoun}
	if kinds := referenceKinds(atoms.References()); !equalKinds(kinds, want) {
		t.Fatalf("reference kinds = %v; want %v", kinds, want)
	}
	if atoms.References()[2].Pronoun != PronounIt || atoms.References()[3].Pronoun != PronounTheir {
		t.Fatalf("pronouns = %v, %v; want it, their", atoms.References()[2].Pronoun, atoms.References()[3].Pronoun)
	}
}

// TestParseEmitsThisSagaSelfReference proves a Saga's "This Saga" self-reference
// is recognized as a ReferenceThisObject so the source permanent backs effects
// like "This Saga deals 4 damage to any target."
func TestParseEmitsThisSagaSelfReference(t *testing.T) {
	t.Parallel()
	references := atomsFor(t, "This Saga deals 4 damage to any target", "").References()
	if len(references) != 1 || references[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %+v; want one this-object reference", references)
	}
}

// TestParseEmitsSelfTypeMarkerReferences proves the permanent type/subtype self
// markers "this Aura", "this Vehicle", and "this Saga" are recognized as
// source-object references (the duration anchor of an O-Ring exile-until-leaves
// clause), the same kind as "this creature"/"this enchantment".
func TestParseEmitsSelfTypeMarkerReferences(t *testing.T) {
	t.Parallel()
	for _, phrase := range []string{"this Aura", "this Vehicle", "this Saga"} {
		source := "until " + phrase + " leaves the battlefield"
		references := atomsFor(t, source, "").References()
		if len(references) != 1 || references[0].Kind != ReferenceThisObject {
			t.Fatalf("%q references = %+v; want one this-object reference", source, references)
		}
	}
}

func TestParseEmitsChosenCardsReference(t *testing.T) {
	t.Parallel()
	source := "return the chosen cards to the battlefield tapped"
	references := atomsFor(t, source, "").References()
	if len(references) != 1 || references[0].Kind != ReferenceChosenCards {
		t.Fatalf("references = %+v; want one chosen-cards reference", references)
	}
	if got := shared.SliceSpan(source, references[0].Span); got != "the chosen cards" {
		t.Fatalf("chosen-cards span = %q; want %q", got, "the chosen cards")
	}
}

func equalKinds(a, b []ReferenceKind) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestParseDurationSubjectIsNotReference proves a source-tied duration subject is
// intentionally not reported as an explicit reference.
func TestParseDurationSubjectIsNotReference(t *testing.T) {
	t.Parallel()
	if refs := atomsFor(t, "for as long as you control Llanowar Elves", "Llanowar Elves").References(); len(refs) != 0 {
		t.Errorf("references = %+v; want none for duration subject", refs)
	}
	if refs := atomsFor(t, "for as long as you control this creature", "").References(); len(refs) != 0 {
		t.Errorf("references = %+v; want none for duration this-object", refs)
	}
}

// TestParseEmptyCardNameSkipsNameScan proves no self-name is recognized without a
// card name in context.
func TestParseEmptyCardNameSkipsNameScan(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "Goblin Guide attacks", "")
	if len(atoms.References()) != 0 || len(atoms.SelfNameSpans()) != 0 {
		t.Errorf("references=%+v spans=%+v; want none without a card name", atoms.References(), atoms.SelfNameSpans())
	}
}
