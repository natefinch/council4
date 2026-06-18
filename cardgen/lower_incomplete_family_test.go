package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// TestUnsupportedEffectFamilyFromTypedContent verifies that an unconsumed
// ability body is attributed to its typed effect family - add-mana, delayed
// one-shot, or multi-effect ordered sequence - and otherwise keeps the generic
// incomplete-lowering reason. The family is derived from typed compiler content
// only, never Oracle wording.
func TestUnsupportedEffectFamilyFromTypedContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content compiler.AbilityContent
		want    string
	}{
		{
			name: "add-mana",
			content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{Kind: compiler.EffectAddMana}},
			},
			want: "unsupported mana symbol",
		},
		{
			name: "delayed one-shot",
			content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{
					Kind:          compiler.EffectSacrifice,
					DelayedTiming: game.DelayedAtBeginningOfNextEndStep,
				}},
			},
			want: "unsupported delayed effect",
		},
		{
			name: "delayed precedence over add-mana",
			content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{
					{Kind: compiler.EffectAddMana},
					{Kind: compiler.EffectSacrifice, DelayedTiming: game.DelayedAtBeginningOfNextEndStep},
				},
			},
			want: "unsupported delayed effect",
		},
		{
			name: "multi-effect ordered sequence",
			content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{
					{Kind: compiler.EffectScry},
					{Kind: compiler.EffectDraw},
				},
			},
			want: "unsupported ordered effect sequence",
		},
		{
			name: "single unknown effect keeps generic",
			content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{Kind: compiler.EffectTransform}},
			},
			want: "incomplete executable lowering",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			summary, _ := unsupportedEffectFamily(test.content)
			if summary != test.want {
				t.Fatalf("summary = %q, want %q", summary, test.want)
			}
		})
	}
}

// TestIncompleteManaAbilityReportsManaSymbol confirms that an add-mana activated
// ability whose reminder text is not fully consumed surfaces the specific
// "unsupported mana symbol" family diagnostic rather than the generic
// incomplete-lowering reason.
func TestIncompleteManaAbilityReportsManaSymbol(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Crawler",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "{T}: Add {C}. ({C} represents colorless mana.)",
		Power:      new("0"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic")
	}
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.Summary == "unsupported mana symbol" {
			found = true
		}
		if diagnostic.Summary == "incomplete executable lowering" {
			t.Fatalf("got generic incomplete diagnostic, want family-specific mana symbol: %#v", diagnostics)
		}
	}
	if !found {
		t.Fatalf("missing unsupported mana symbol diagnostic: %#v", diagnostics)
	}
}
