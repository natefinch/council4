package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestReferenceSubjectReachedCardMovementAdapter(t *testing.T) {
	for _, text := range []string{
		"{1}: Look at the top card of your library. Put it onto the battlefield. Put it into your hand.",
		"{1}: Look at the top card of your library. You may put it onto the battlefield. Put it into your hand.",
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := parser.Parse(text, parser.Context{CardName: "Reference Capability"})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}

			compilation, diagnostics := compiler.Compile(document, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			ctx := contentCtx{
				content: content, text: text, span: document.Abilities[0].Span,
				enclosingKind: compiler.AbilityActivated,
			}
			if _, diagnostic := lowerOrderedEffectSequence("Reference Capability", ctx, &document.Abilities[0]); diagnostic != nil {
				t.Fatalf("normal reached-card movement: %#v", diagnostic)
			}
			assertCardPaths(t, &ScryfallCard{
				Name: "Reference Capability", Layout: "normal", TypeLine: "Artifact", OracleText: text,
			}, "ActivatedAbilities[0].Content.Modes[0].Sequence[2].Primitive.(game.MovePermanent).Destination = zone.Hand")
		})
	}
}
