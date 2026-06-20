package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileEffectFollowsTypedParserSyntax(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse("irrelevant", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	sentence := &document.Abilities[0].Sentences[0]
	sentence.Targets = []parser.TargetSyntax{{
		Span:        sentence.Span,
		Text:        "irrelevant",
		Cardinality: parser.TargetCardinalitySyntax{Min: 0, Max: 2},
		Selection: parser.SelectionSyntax{
			Span:       sentence.Span,
			Text:       "irrelevant",
			Kind:       parser.SelectionCreature,
			Controller: parser.SelectionControllerOpponent,
			Supertypes: []parser.Supertype{parser.SupertypeLegendary},
		},
	}}
	sentence.Effects = []parser.EffectSyntax{{
		Kind:           parser.EffectReturn,
		Span:           sentence.Span,
		ClauseSpan:     sentence.Span,
		VerbSpan:       sentence.Span,
		Text:           "irrelevant",
		Targets:        sentence.Targets,
		SubjectTargets: sentence.Targets,
		References: []parser.Reference{{
			Kind: parser.ReferenceThatObject,
			Span: sentence.Span,
		}},
		Duration:     parser.EffectDurationUntilYourNextTurn,
		Selection:    sentence.Targets[0].Selection,
		Amount:       parser.EffectAmountSyntax{Value: 3, Known: true},
		CounterKind:  counter.Charge,
		CounterKnown: true,
		FromZone:     zone.Graveyard,
		ToZone:       zone.Hand,
		Mana: parser.EffectManaSyntax{
			Symbols: []string{"{G}", "{W}"},
			Choice:  true,
		},
		Replacement: parser.EffectReplacementSyntax{
			Kind:   parser.EffectReplacementThatMuchPlus,
			Amount: 2,
		},
	}}

	compilation, compileDiagnostics := Compile(document, Context{})
	if len(compileDiagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", compileDiagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 1 || content.Effects[0].Kind != EffectReturn ||
		content.Effects[0].Duration != DurationUntilYourNextTurn ||
		content.Effects[0].Amount.Value != 3 ||
		content.Effects[0].CounterKind != counter.Charge ||
		content.Effects[0].FromZone != zone.Graveyard ||
		content.Effects[0].ToZone != zone.Hand ||
		!content.Effects[0].Mana.Choice ||
		len(content.Effects[0].Mana.Symbols) != 2 ||
		content.Effects[0].Replacement.Kind != parser.EffectReplacementThatMuchPlus ||
		content.Effects[0].Replacement.Amount != 2 ||
		content.Effects[0].ClauseSpan != sentence.Span ||
		len(content.Effects[0].Targets) != 1 ||
		len(content.Effects[0].SubjectTargets) != 1 ||
		len(content.Effects[0].References) != 1 {
		t.Fatalf("effect = %#v", content.Effects)
	}
	if len(content.Targets) != 1 ||
		content.Targets[0].Cardinality != (TargetCardinality{Min: 0, Max: 2}) ||
		content.Targets[0].Selector.Kind != SelectorCreature ||
		content.Targets[0].Selector.Controller != ControllerOpponent ||
		len(content.Targets[0].Selector.Supertypes()) != 1 ||
		content.Targets[0].Selector.Supertypes()[0] != types.Legendary {
		t.Fatalf("targets = %#v", content.Targets)
	}
}

func TestCompileModalEffectOwnershipReceivesReferenceBindings(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Choose one —\n• Return target creature to its owner's hand, then draw a card.\n• Draw two cards.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Modes[0].Content.Effects
	if len(effects) != 2 || len(effects[0].References) != 1 ||
		effects[0].References[0].Binding != ReferenceBindingTarget {
		t.Fatalf("modal effects = %#v", effects)
	}
}

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
func TestCompileKeywordFollowsTypedParserSyntax(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "irrelevant")
	parameter := parser.NewProtectionKeywordParameter(tokens[0].Span, "subtypes:Dragon", parser.ProtectionParameter{
		FromSubtypes: []types.Sub{types.Dragon},
	})
	atoms := parser.NewAtoms(parser.WithKeywords(parser.Keyword{
		Kind:      parser.KeywordProtection,
		NameSpan:  tokens[0].Span,
		Span:      tokens[0].Span,
		Text:      "irrelevant",
		Parameter: parameter,
	}))
	keywords := compileKeywords(atoms.KeywordsWithin(tokens))
	if len(keywords) != 1 ||
		keywords[0].Kind != parser.KeywordProtection ||
		keywords[0].Name != "Protection" ||
		!keywords[0].ProtectionKnown ||
		len(keywords[0].Protection.FromSubtypes) != 1 ||
		keywords[0].Protection.FromSubtypes[0] != types.Dragon {
		t.Fatalf("compiled keywords = %+v; want typed Protection from Dragon", keywords)
	}

	if keywords[0].Text != "irrelevant" {
		t.Fatalf("keyword source metadata = %q; want irrelevant", keywords[0].Text)
	}
}

func TestCompileKeywordParameterShapesFollowTypedParserSyntax(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "irrelevant")
	tests := []struct {
		name      string
		kind      parser.KeywordKind
		parameter parser.KeywordParameter
		check     func(CompiledKeyword) bool
	}{
		{
			name:      "mana",
			kind:      parser.KeywordWard,
			parameter: parser.NewManaKeywordParameter(tokens[0].Span, cost.Mana{cost.U}),
			check: func(keyword CompiledKeyword) bool {
				return keyword.ParameterKind == parser.KeywordParameterManaCost &&
					len(keyword.ManaCost) == 1 && keyword.ManaCost[0] == cost.U
			},
		},
		{
			name:      "integer",
			kind:      parser.KeywordToxic,
			parameter: parser.NewIntegerKeywordParameter(tokens[0].Span, 7),
			check: func(keyword CompiledKeyword) bool {
				return keyword.ParameterKind == parser.KeywordParameterInteger && keyword.Integer == 7
			},
		},
		{
			name:      "enchant target",
			kind:      parser.KeywordEnchant,
			parameter: parser.NewEnchantTargetKeywordParameter(tokens[0].Span, parser.ObjectNounPlayer),
			check: func(keyword CompiledKeyword) bool {
				return keyword.ParameterKind == parser.KeywordParameterEnchantTarget &&
					keyword.EnchantTarget == parser.ObjectNounPlayer
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			atoms := parser.NewAtoms(parser.WithKeywords(parser.Keyword{
				Kind:      test.kind,
				NameSpan:  tokens[0].Span,
				Span:      tokens[0].Span,
				Text:      "irrelevant",
				Parameter: test.parameter,
			}))
			keywords := compileKeywords(atoms.KeywordsWithin(tokens))
			if len(keywords) != 1 || keywords[0].Kind != test.kind || !test.check(keywords[0]) {
				t.Fatalf("compiled keywords = %+v; want typed %v", keywords, test.kind)
			}
		})
	}
}

func TestCompileCumulativeUpkeepFollowsTypedParserSyntax(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "not cumulative upkeep")
	parameter := parser.NewManaKeywordParameter(tokens[0].Span, cost.Mana{cost.O(1), cost.U})
	atoms := parser.NewAtoms(parser.WithKeywords(parser.Keyword{
		Kind:      parser.KeywordCumulativeUpkeep,
		NameSpan:  tokens[0].Span,
		Span:      tokens[0].Span,
		Text:      "not cumulative upkeep",
		Parameter: parameter,
	}))
	keywords := compileKeywords(atoms.KeywordsWithin(tokens))
	if len(keywords) != 1 ||
		keywords[0].Kind != parser.KeywordCumulativeUpkeep ||
		keywords[0].Name != "Cumulative upkeep" ||
		keywords[0].Span != tokens[0].Span ||
		keywords[0].ParameterKind != parser.KeywordParameterManaCost ||
		!slices.Equal(keywords[0].ManaCost, cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("compiled keywords = %+v; want typed cumulative upkeep", keywords)
	}
}

func TestCompileReferencesFollowsTypedAtoms(t *testing.T) {
	t.Parallel()
	tokens := compilerTokens(t, "Mistform Ultimus attacks")
	atoms := parser.NewAtoms(parser.WithReferences(parser.Reference{
		Kind:    parser.ReferencePronoun,
		Pronoun: parser.PronounTheir,
		Span:    shared.SpanOf(tokens[0:2]),
		Tokens:  tokens[0:2],
	}))
	references := compileTypedReferences(atoms.ReferencesWithin(tokens))
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
	if refs := compileTypedReferences(atoms.ReferencesWithin(tokens[2:])); len(refs) != 0 {
		t.Errorf("references over disjoint tokens = %+v; want none", refs)
	}
}

func TestCompileReferenceKindMapping(t *testing.T) {
	t.Parallel()
	cases := map[parser.ReferenceKind]ReferenceKind{
		parser.ReferenceSelfName:   ReferenceSelfName,
		parser.ReferenceThisObject: ReferenceThisObject,
		parser.ReferenceThatObject: ReferenceThatObject,
		parser.ReferenceThatPlayer: ReferenceThatPlayer,
		parser.ReferencePronoun:    ReferencePronoun,
		parser.ReferenceUnknown:    ReferenceUnknown,
	}
	for atom, want := range cases {
		if got := compileReferenceKind(atom); got != want {
			t.Errorf("compileReferenceKind(%v) = %v; want %v", atom, got, want)
		}
	}
}
