package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileUmbraArmorKeyword(t *testing.T) {
	t.Parallel()
	source := "Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)"
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Test Umbra"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Keywords) != 1 {
		t.Fatalf("abilities = %#v, want one keyword ability", compilation.Abilities)
	}
	keyword := compilation.Abilities[0].Content.Keywords[0]
	if keyword.Kind != parser.KeywordUmbraArmor || keyword.ParameterKind != parser.KeywordParameterNone {
		t.Fatalf("keyword = %#v, want typed Umbra armor with no parameter", keyword)
	}
}
