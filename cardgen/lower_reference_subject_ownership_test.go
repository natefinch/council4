package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestReferenceSubjectAdapterRefusesMissingIncompatibleAndForeignProof(t *testing.T) {
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
	for _, kind := range []string{"missing", "incompatible", "foreign"} {
		t.Run(kind, func(t *testing.T) {
			document, content := compile()
			switch kind {
			case "foreign":
				_, other := compile()
				content.References[0] = other.References[0]
			case "incompatible":
				content.References[0].Binding = compiler.ReferenceBindingSource
			default:
				content.References[0].Subject = compiler.ReferenceSubjectProof{}
			}
			ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span,
				enclosingKind: compiler.AbilityActivated}
			lowered, diagnostic := lowerContent("Proof Ownership", ctx, &document.Abilities[0])
			if diagnostic == nil || diagnostic.Summary != "unsupported reference subject" || !reflect.DeepEqual(lowered, game.AbilityContent{}) {
				t.Fatalf("unowned subject admitted or primary rejection lost: %+v / %+v", lowered, diagnostic)
			}
			if diagnostic.Span != ctx.span || diagnostic.Detail != "the body contains a missing, incompatible, or foreign subject proof" {
				t.Fatalf("proof rejection metadata changed: %+v", diagnostic)
			}
		})
	}
}

func TestReferenceSubjectDiagnosticProbeCannotAdmitForeignDamage(t *testing.T) {
	const text = "Proof Damage deals 1 damage to any target."
	compile := func() (parser.Document, compiler.AbilityContent) {
		document, diagnostics := parser.Parse(text, parser.Context{CardName: "Proof Damage", InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return document, compilation.Abilities[0].Content
	}
	document, content := compile()
	_, foreign := compile()
	content.References[0] = foreign.References[0]
	ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span, enclosingKind: compiler.AbilitySpell}
	if accepted, diagnostic := lowerContentDispatch("Proof Damage", ctx, &document.Abilities[0]); diagnostic != nil || len(accepted.Modes) != 1 {
		t.Fatalf("fixture does not exercise otherwise accepted dispatch: %+v / %+v", accepted, diagnostic)
	}
	lowered, diagnostic := lowerContent("Proof Damage", ctx, &document.Abilities[0])
	if !reflect.DeepEqual(lowered, game.AbilityContent{}) || diagnostic == nil || diagnostic.Summary != "unsupported reference subject" {
		t.Fatalf("diagnostic probe admitted foreign content: %+v / %+v", lowered, diagnostic)
	}
}
