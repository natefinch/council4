package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// These tests construct typed parser atoms over deliberately irrelevant source
// text and assert that the compiler's lowered meaning follows the typed atom,
// not the token spelling. The compiler no longer recognizes these atoms from
// text.

func compilerTokens(t *testing.T, source string) []shared.Token {
	t.Helper()
	document, diagnostics := parser.Parse(source, parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("parse %q diagnostics: %v", source, diagnostics)
	}
	if len(document.Abilities) == 0 {
		t.Fatalf("parse %q produced no abilities", source)
	}
	return document.Abilities[0].Tokens
}

// TestCompileFromZoneFollowsTypedAtom: the spelling says "graveyard" but the
// emitted atom says Exile, so the compiler must return Exile.
func TestCompileFromZoneFollowsTypedAtom(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "from your graveyard")
	atoms := parser.NewAtoms(parser.WithZones(parser.ZoneAtom{
		Zone: zone.Exile,
		Role: parser.ZoneRoleFrom,
		Span: tokens[0].Span,
	}))
	if got := compileFromZone(tokens, atoms); got != zone.Exile {
		t.Errorf("compileFromZone = %v; want %v (typed atom, not spelling)", got, zone.Exile)
	}
	// With no emitted zone atom the compiler reports no zone regardless of text.
	if got := compileFromZone(tokens, parser.Atoms{}); got != zone.None {
		t.Errorf("compileFromZone(no atom) = %v; want none", got)
	}
}

// TestCounterKindWordFollowsTypedAtom: "lorwyn" is not a counter name, but the
// emitted atom types it as a Charge counter.
func TestCounterKindWordFollowsTypedAtom(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "a lorwyn counter")
	atoms := parser.NewAtoms(parser.WithCounters(parser.CounterAtom{
		Kind: counter.Charge,
		Span: tokens[1].Span, // span over "lorwyn"
	}))
	kind, ok := counterKindWord(tokens, atoms)
	if !ok || kind != counter.Charge {
		t.Errorf("counterKindWord = %v, %v; want charge, true (typed atom)", kind, ok)
	}
	// Without the typed atom the counter kind is unknown even though the text is
	// unchanged.
	if _, ok := counterKindWord(tokens, parser.Atoms{}); ok {
		t.Error("counterKindWord(no atom) = true; want false")
	}
}

// TestNumberWordFollowsTypedAtom: a cardinal value comes from the emitted atom,
// not from the word's spelling, and the compiler keeps its <=4 range policy.
func TestNumberWordFollowsTypedAtom(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "zzz")
	three := parser.NewAtoms(parser.WithCardinals(parser.CardinalAtom{Value: 3, Span: tokens[0].Span}))
	if got := numberWord(tokens[0], three); got != 3 {
		t.Errorf("numberWord = %d; want 3 (typed atom)", got)
	}
	// Values above the compiler's conservative cap are rejected.
	seven := parser.NewAtoms(parser.WithCardinals(parser.CardinalAtom{Value: 7, Span: tokens[0].Span}))
	if got := numberWord(tokens[0], seven); got != 0 {
		t.Errorf("numberWord(7) = %d; want 0 (capped)", got)
	}
	// Integer literals are read structurally, not via atoms.
	intTokens := compilerTokens(t, "5")
	if got := numberWord(intTokens[0], parser.Atoms{}); got != 5 {
		t.Errorf("numberWord(integer) = %d; want 5", got)
	}
}

// TestCompileProtectionParameterFollowsTypedSubtype: "from gnomes" is not a known
// subtype, but the emitted atom types it as Dragon.
func TestCompileProtectionParameterFollowsTypedSubtype(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "from gnomes")
	atoms := parser.NewAtoms(parser.WithSubtypes(parser.SubtypeAtom{
		Identity: types.Dragon,
		Span:     tokens[1].Span, // span over "gnomes"
	}))
	parameter := compileProtectionParameter(tokens, 0, atoms)
	if !parameter.ok || parameter.text != "subtypes:Dragon" {
		t.Errorf("compileProtectionParameter = %q, %v; want subtypes:Dragon, true", parameter.text, parameter.ok)
	}
}

func TestProtectionSubtypePreservesSupportedFamilies(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "from gnomes")
	for _, test := range []struct {
		name string
		sub  types.Sub
		ok   bool
	}{
		{name: "creature", sub: types.Dragon, ok: true},
		{name: "land", sub: types.Forest, ok: true},
		{name: "artifact", sub: types.Equipment},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			atoms := parser.NewAtoms(parser.WithSubtypes(parser.SubtypeAtom{
				Identity: test.sub,
				Span:     tokens[1].Span,
			}))
			got, ok := protectionSubtype(tokens[1], atoms)
			if ok != test.ok || ok && got != test.sub {
				t.Fatalf("protectionSubtype = %q, %v; want %q, %v", got, ok, test.sub, test.ok)
			}
		})
	}
}

// TestCompileReferencesFollowsTypedAtoms: the lowered references follow the typed
// reference atoms, including their span and kind, irrespective of token text.
func TestCompileReferencesFollowsTypedAtoms(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "Mistform Ultimus attacks")
	atoms := parser.NewAtoms(parser.WithReferences(parser.Reference{
		Kind:    parser.ReferencePronoun,
		Pronoun: parser.PronounTheir,
		Span:    shared.SpanOf(tokens[0:2]),
		Tokens:  tokens[0:2],
	}))
	references := compileReferences(tokens, atoms)
	if len(references) != 1 ||
		references[0].Kind != ReferencePronoun ||
		references[0].Pronoun != ReferencePronounTheir {
		t.Fatalf("references = %+v; want one their-pronoun", references)
	}
	if references[0].Span != shared.SpanOf(tokens[0:2]) {
		t.Errorf("reference span = %+v; want %+v", references[0].Span, shared.SpanOf(tokens[0:2]))
	}
	// A reference whose first token is outside the supplied selection is not
	// reported, letting callers consume references over a token subset.
	if refs := compileReferences(tokens[2:], atoms); len(refs) != 0 {
		t.Errorf("references over disjoint tokens = %+v; want none", refs)
	}
}

func TestCompileReferenceKindMapping(t *testing.T) {
	t.Parallel()
	cases := map[parser.ReferenceKind]ReferenceKind{
		parser.ReferenceSelfName:   ReferenceSelfName,
		parser.ReferenceThisObject: ReferenceThisObject,
		parser.ReferenceThatObject: ReferenceThatObject,
		parser.ReferencePronoun:    ReferencePronoun,
		parser.ReferenceUnknown:    ReferenceUnknown,
	}
	for atom, want := range cases {
		if got := compileReferenceKind(atom); got != want {
			t.Errorf("compileReferenceKind(%v) = %v; want %v", atom, got, want)
		}
	}
}

func TestCompileSelectorFollowsTypedNounAndModifiers(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "sparkly zed")
	atoms := parser.NewAtoms(
		parser.WithObjectNouns(parser.ObjectNounAtom{Noun: parser.ObjectNounCreature, Span: tokens[1].Span}),
		parser.WithSelectionFlags(parser.SelectionFlagAtom{Flag: parser.SelectionFlagTapped, Span: tokens[0].Span}),
	)
	selector := compileSelector(tokens, atoms)
	if selector.Kind != SelectorCreature || !selector.Tapped {
		t.Fatalf("selector = %+v; want typed creature and tapped", selector)
	}
	if selector := compileSelector(tokens, parser.Atoms{}); selector.Kind != SelectorUnknown || selector.Tapped {
		t.Fatalf("selector without atoms = %+v; want unknown untapped", selector)
	}
}

func TestTriggerSelfSubjectFollowsTypedReference(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "this creature enters")
	syntax := newTriggerEventSyntax("gibberish enters", tokens, parser.NewAtoms(parser.WithReferences(parser.Reference{
		Kind:   parser.ReferenceThisObject,
		Span:   shared.SpanOf(tokens[:2]),
		Tokens: tokens[:2],
	})))
	if !syntax.selfSubject("this creature", selfEnterSubjectSlots, true) {
		t.Fatal("typed this-object reference was not accepted as source subject")
	}
	without := newTriggerEventSyntax("this creature enters", tokens, parser.Atoms{})
	if without.selfSubject("this creature", selfEnterSubjectSlots, true) {
		t.Fatal("source subject recognized without typed reference atom")
	}
}
