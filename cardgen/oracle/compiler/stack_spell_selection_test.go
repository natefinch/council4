package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileStackSpellSelectionOwnership(t *testing.T) {
	for _, text := range []string{
		"Counter target blue enchantment spell you don't control with mana value 3 or less.",
		"Exile any number of target blue enchantment spells you don't control with mana value 3 or less.",
	} {
		t.Run(text, func(t *testing.T) {
			compilation, diagnostics := compileSource(text, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}

			content := compilation.Abilities[0].Content
			if len(content.Targets) != 1 || len(content.Effects) != 1 {
				t.Fatalf("content = %#v", content)
			}
			effect := content.Effects[0]
			target := content.Targets[0]
			if !effect.Exact || !target.Exact || effect.Amount.Known || effect.Amount.RangeKnown ||
				effect.Amount.DynamicKind != DynamicAmountNone {
				t.Fatalf("target-owned qualifiers polluted effect: %#v; target = %#v", effect, target)
			}
			sel := target.Selector
			if sel.Kind != SelectorSpell || sel.Controller != ControllerNotYou ||
				!slices.Equal(sel.RequiredTypesAny(), []types.Card{types.Enchantment}) ||
				!slices.Equal(sel.ColorsAny(), []color.Color{color.Blue}) ||
				!sel.MatchManaValue || sel.ManaValue.Value != 3 {
				t.Fatalf("selector = %#v, atoms = %#v", sel, sel.atoms)
			}
		})
	}
}

func TestCompileStackSpellUnionOwnership(t *testing.T) {
	text := "Exile any number of target artifact or enchantment spells an opponent controls with mana value 3 or less."
	compilation, diagnostics := compileSource(text, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	effect := content.Effects[0]
	target := content.Targets[0]
	if !effect.Exact || !target.Exact || effect.Amount.Known || effect.Amount.RangeKnown ||
		effect.Context != parser.EffectContextController || len(content.References) != 0 ||
		len(content.Keywords) != 0 || len(content.Conditions) != 0 {
		t.Fatalf("target-owned qualifiers polluted content: effect=%#v target=%#v references=%#v keywords=%#v conditions=%#v",
			effect, target, content.References, content.Keywords, content.Conditions)
	}
	if !slices.Equal(target.Selector.RequiredTypesAny(), []types.Card{types.Artifact, types.Enchantment}) ||
		target.Selector.Controller != ControllerOpponent {
		t.Fatalf("selector = %#v", target.Selector)
	}
}
