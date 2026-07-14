package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func graveyardGrantDeclaration(t *testing.T, compilation Compilation) *StaticGraveyardKeywordGrantDeclaration {
	t.Helper()
	for _, ability := range compilation.Abilities {
		if ability.Static == nil {
			continue
		}
		for _, declaration := range ability.Static.Declarations {
			if declaration.GraveyardGrant != nil {
				return declaration.GraveyardGrant
			}
		}
	}
	t.Fatal("compilation produced no graveyard keyword-grant declaration")
	return nil
}

// TestCompileGraveyardEscapeGrantDeclaration proves the compiler recognizes
// Underworld Breach's escape grant, resolving the granted keyword to Escape,
// the nonland filter, and the computed escape cost payload, with no diagnostics.
func TestCompileGraveyardEscapeGrantDeclaration(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	grant := graveyardGrantDeclaration(t, compilation)
	if grant.Keyword.Kind != parser.KeywordEscape {
		t.Fatalf("granted keyword = %v, want Escape", grant.Keyword.Kind)
	}
	if grant.Filter != parser.StaticDeclarationCardFilterNonland {
		t.Fatalf("filter = %v, want nonland", grant.Filter)
	}
	if grant.DuringControllerTurn {
		t.Fatal("escape grant is not restricted to the controller's turn")
	}
	if grant.EscapeCost == nil {
		t.Fatal("EscapeCost = nil, want computed escape cost")
	}
	if !grant.EscapeCost.UseCardManaCost {
		t.Fatal("UseCardManaCost = false, want true")
	}
	if grant.EscapeCost.ExileOtherCount != 3 {
		t.Fatalf("ExileOtherCount = %d, want 3", grant.EscapeCost.ExileOtherCount)
	}
}
