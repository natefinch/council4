package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileLinkedSearchRiderReference(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	content := compilation.Abilities[0].Content
	if len(content.Conditions) != 1 || !content.Conditions[0].Resolving {
		t.Fatalf("conditions = %#v, want resolving condition", content.Conditions)
	}
	if len(content.References) == 0 {
		t.Fatal("missing compiled references")
	}
	ref := content.References[len(content.References)-1]
	if ref.Kind != ReferenceThatObject ||
		ref.Binding != ReferenceBindingPriorInstructionResult ||
		ref.PriorInstruction != 0 {
		t.Fatalf("rider reference = %#v, want prior search result", ref)
	}
}

func TestCompileLibraryTopTutorIsTypedAndReferencesSearchResult(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 4 {
		t.Fatalf("effects = %#v, want search/reveal/shuffle/put", content.Effects)
	}
	search := content.Effects[0]
	if !search.Exact || search.SearchDestination != parser.EffectDestinationTop ||
		!slices.Equal(search.Selector.RequiredTypesAny(), []types.Card{types.Artifact, types.Enchantment}) {
		t.Fatalf("search = %#v, want typed top artifact-or-enchantment search", search)
	}
	if len(content.References) != 2 {
		t.Fatalf("references = %#v, want reveal and put references", content.References)
	}
	for _, ref := range content.References {
		if ref.Binding != ReferenceBindingPriorInstructionResult || ref.PriorInstruction != 0 {
			t.Fatalf("reference = %#v, want prior search result", ref)
		}
	}
}

func TestCompileSpellTypeTutorKeepsRequiredCardType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       []types.Card
	}{
		{
			name:       "instant or sorcery union",
			oracleText: "Search your library for an instant or sorcery card, reveal it, then shuffle and put that card on top.",
			want:       []types.Card{types.Instant, types.Sorcery},
		},
		{
			name:       "single sorcery type",
			oracleText: "Search your library for a sorcery card, reveal it, then shuffle and put that card on top.",
			want:       []types.Card{types.Sorcery},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.oracleText, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			search := compilation.Abilities[0].Content.Effects[0]
			if !search.Exact || search.SearchDestination != parser.EffectDestinationTop ||
				!slices.Equal(search.Selector.RequiredTypesAny(), test.want) {
				t.Fatalf("search = %#v, want typed top search with %v", search, test.want)
			}
		})
	}
}
