package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileRecipientControlsSelectorCapturesQualifier verifies the "who
// controls an artifact or enchantment" per-member group qualifier (Fade from
// History) compiles onto the create effect's RecipientControlsSelector, keeping
// the base each-player context and describing the controlled permanent's
// artifact-or-enchantment union with an any controller.
func TestCompileRecipientControlsSelectorCapturesQualifier(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Each player who controls an artifact or enchantment creates a 2/2 green Bear creature token. Then destroy all artifacts and enchantments.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 2 {
		t.Fatalf("effects = %#v, want create then destroy", content.Effects)
	}
	create := content.Effects[0]
	if create.Kind != EffectCreate {
		t.Fatalf("effect[0] kind = %v, want EffectCreate", create.Kind)
	}
	if create.Context != parser.EffectContextEachPlayer {
		t.Fatalf("effect[0] context = %v, want the each-player base group", create.Context)
	}
	if create.RecipientControlsSelector == nil {
		t.Fatal("RecipientControlsSelector = nil, want the compiled qualifier")
	}
	got := create.RecipientControlsSelector.RequiredTypesAny()
	want := []types.Card{types.Artifact, types.Enchantment}
	if len(got) != len(want) {
		t.Fatalf("qualifier RequiredTypesAny = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("qualifier RequiredTypesAny = %v, want %v", got, want)
		}
	}
	if create.RecipientControlsSelector.Controller != ControllerAny {
		t.Fatalf("qualifier Controller = %v, want Any (per-member control checked at runtime)", create.RecipientControlsSelector.Controller)
	}
	// An unqualified group recipient carries no selector.
	unqualified, diagnostics := compileSource(
		"Each player creates a 2/2 green Bear creature token.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if selector := unqualified.Abilities[0].Content.Effects[0].RecipientControlsSelector; selector != nil {
		t.Fatalf("unqualified RecipientControlsSelector = %#v, want nil", selector)
	}
}
