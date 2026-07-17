package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileXBoundedHandFreeCastFromTypedNodes(t *testing.T) {
	t.Parallel()
	sentences := []parser.Sentence{{
		Effects: []parser.EffectSyntax{{
			Kind:                      parser.EffectCast,
			Context:                   parser.EffectContextController,
			FromZone:                  zone.Hand,
			CastWithoutPayingManaCost: true,
			Optional:                  true,
			Selection: parser.SelectionSyntax{
				Kind:           parser.SelectionSpell,
				MatchManaValue: true,
				ManaValueX:     true,
				ManaValue:      compare.Int{Op: compare.LessOrEqual},
			},
		}},
	}}
	effects := compileEffects(sentences)
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	got := effects[0]
	if got.Kind != EffectCast ||
		!got.Optional ||
		!got.CastWithoutPayingManaCost ||
		got.FromZone != zone.Hand ||
		!got.Selector.MatchManaValue ||
		!got.Selector.ManaValueX ||
		got.Selector.ManaValue.Op != compare.LessOrEqual {
		t.Fatalf("compiled effect = %#v", got)
	}
}
