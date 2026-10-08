package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestDestroyedTargetDamageRefusesUnownedCharacteristicProof(t *testing.T) {
	const text = "Destroy target nonblack creature. It can't be regenerated. Proof Spell deals damage equal to that creature's power to the creature's controller."
	compile := func() (parser.Document, compiler.AbilityContent) {
		t.Helper()
		document, diagnostics := parser.Parse(text, parser.Context{CardName: "Proof Spell", InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
			t.Fatal(diagnostics)
		}
		return document, compilation.Abilities[0].Content
	}
	for _, corruption := range []string{"missing", "foreign", "binding", "node"} {
		t.Run(corruption, func(t *testing.T) {
			document, content := compile()
			index := -1
			for i, ref := range content.References {
				if ref.Binding == compiler.ReferenceBindingTarget && ref.ExactDamageTargetCharacteristic(content.Effects[1].Amount) {
					index = i
				}
			}
			if index < 0 {
				t.Fatal("fixture lacks an exact damage target-characteristic proof")
			}
			ref := &content.References[index]
			switch corruption {
			case "missing":
				ref.Subject = compiler.ReferenceSubjectProof{}
			case "foreign":
				_, foreign := compile()
				*ref = foreign.References[index]
			case "binding":
				ref.Binding = compiler.ReferenceBindingSource
			case "node":
				ref.NodeID++
			default:
				t.Fatal("unknown corruption")
			}
			ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span, enclosingKind: compiler.AbilitySpell}
			lowered, diagnostic := lowerContent("Proof Spell", ctx, &document.Abilities[0])
			if diagnostic == nil || diagnostic.Summary != "unsupported reference subject" ||
				!reflect.DeepEqual(lowered, game.AbilityContent{}) {
				t.Fatalf("unowned target characteristic admitted: %+v / %+v", lowered, diagnostic)
			}
		})
	}
}
