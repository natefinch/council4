package cardgen

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRenderActualCardQuantityEmptySelection(t *testing.T) {
	for _, selected := range []bool{false, true} {
		dynamic := game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "actual"}
		if selected {
			dynamic.Selection = &game.Selection{}
		}
		text, err := (Renderer{}).renderQuantity(newRenderCtx(), game.Dynamic(dynamic))
		if err != nil {
			t.Fatal(err)
		}
		expression, err := goparser.ParseExprFrom(token.NewFileSet(), "", text, 0)
		if err != nil {
			t.Fatal(err)
		}
		call, ok := expression.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			t.Fatalf("rendered quantity is not Dynamic(amount): %s", text)
		}
		literal, ok := call.Args[0].(*ast.CompositeLit)
		if !ok {
			t.Fatalf("rendered amount is not a typed literal: %s", text)
		}
		found := false
		for _, field := range literal.Elts {
			keyValue, ok := field.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := keyValue.Key.(*ast.Ident)
			if !ok || key.Name != "Selection" {
				continue
			}
			address, ok := keyValue.Value.(*ast.UnaryExpr)
			if !ok || address.Op != token.AND {
				t.Fatal("selection field did not retain nonnil pointer identity")
			}
			value, ok := address.X.(*ast.CompositeLit)
			if !ok || len(value.Elts) != 0 {
				t.Fatal("bare-card quantity no longer emits an empty selection")
			}
			found = true
		}
		if found != selected {
			t.Fatalf("rendered snapshot marker=%t, want=%t: %s", found, selected, text)
		}
	}
}
