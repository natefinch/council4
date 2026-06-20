package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileOtawaraChannelShell(t *testing.T) {
	t.Parallel()
	source := "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Otawara, Soaring City"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.AbilityWord != "Channel" || ability.ActivationZone != zone.Hand {
		t.Fatalf("ability word/zone = %q/%v, want Channel/hand", ability.AbilityWord, ability.ActivationZone)
	}
	if ability.Cost == nil || len(ability.Cost.Components) != 2 ||
		ability.Cost.Components[1].Kind != CostDiscard ||
		!ability.Cost.Components[1].SourceSelf ||
		ability.Cost.Components[1].SourceZone != zone.Hand {
		t.Fatalf("cost = %#v, want mana plus discard source from hand", ability.Cost)
	}
	reduction := ability.ActivationCostReduction
	if reduction == nil || reduction.PerObjectReduction != 1 ||
		reduction.Amount.DynamicKind != DynamicAmountCount ||
		reduction.Amount.Multiplier != 1 {
		t.Fatalf("reduction = %#v", reduction)
	}
}
