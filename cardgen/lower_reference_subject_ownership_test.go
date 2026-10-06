package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestReferenceSubjectAdapterRefusesMissingAndForeignProof(t *testing.T) {
	text := "{1}: Reveal the top card of your library. If it's a creature card, draw a card."
	compile := func() (parser.Document, compiler.AbilityContent) {
		document, diagnostics := parser.Parse(text, parser.Context{CardName: "Proof Ownership"})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return document, compilation.Abilities[0].Content
	}
	for _, foreign := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "foreign"}[foreign], func(t *testing.T) {
			document, content := compile()
			if foreign {
				_, other := compile()
				content.References[0] = other.References[0]
			} else {
				content.References[0].Subject = compiler.ReferenceSubjectProof{}
			}
			ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span,
				enclosingKind: compiler.AbilityActivated}
			if _, diagnostic := lowerContent("Proof Ownership", ctx, &document.Abilities[0]); diagnostic == nil {
				t.Fatal("a lowering caller repaired or accepted an unowned subject")
			}
		})
	}
}
