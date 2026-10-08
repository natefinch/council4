package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestReferenceSubjectRejectionRetainsNestedBlockerMetadata(t *testing.T) {
	const text = "Target creature you control fights target creature the opponent to your left controls. Then that player may copy this spell and may choose new targets for the copy."
	document, diagnostics := parser.Parse(text, parser.Context{CardName: "Proof Blockers", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
		t.Fatalf("compile = %+v / %+v", compilation, diagnostics)
	}
	ctx := contentCtx{content: compilation.Abilities[0].Content, text: text,
		span: document.Abilities[0].Span, enclosingKind: compiler.AbilitySpell}
	if contentSubjectProofsOwned(ctx.content) {
		t.Fatal("fixture no longer exercises proof rejection")
	}
	_, blockers := lowerContentDispatch("Proof Blockers", ctx, &document.Abilities[0])
	if blockers == nil {
		t.Fatal("fixture lost its independent blockers")
	}
	content, rejection := lowerContent("Proof Blockers", ctx, &document.Abilities[0])
	primary := contentDiagnostic(ctx, "unsupported reference subject",
		"the body contains a missing, incompatible, or foreign subject proof")
	primary.Additional = []shared.Diagnostic{*blockers}
	if !reflect.DeepEqual(content, game.AbilityContent{}) || !reflect.DeepEqual(rejection, primary) {
		t.Fatalf("proof rejection admitted content or lost primary/nested metadata/order: %+v / %+v", content, rejection)
	}
	flat := flattenAdditionalReasons([]shared.Diagnostic{*rejection})
	want := append(flattenAdditionalReasons([]shared.Diagnostic{*contentDiagnostic(ctx, "unsupported reference subject",
		"the body contains a missing, incompatible, or foreign subject proof")}),
		flattenAdditionalReasons([]shared.Diagnostic{*blockers})...)
	if !reflect.DeepEqual(flat, want) || !hasDiagnosticSummary(flat, "unsupported optional effect") ||
		!hasDiagnosticSummary(flat, "unsupported ordered effect sequence") {
		t.Fatalf("proof rejection lost flat blocker metadata/order: %+v", flat)
	}
}
