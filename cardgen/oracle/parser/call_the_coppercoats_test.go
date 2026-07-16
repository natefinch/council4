package parser

import (
	"slices"
	"testing"
)

const callTheCoppercoatsOracle = "Strive — This spell costs {1}{W} more to cast for each target beyond the first.\nChoose any number of target opponents. Create X 1/1 white Human Soldier creature tokens, where X is the number of creatures those opponents control."

func TestParseCallTheCoppercoatsMechanics(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(callTheCoppercoatsOracle, Context{CardName: "Call the Coppercoats", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	strive := document.Abilities[0].Sentences[0].Effects[0]
	if !strive.SourceSpellCostIncreasePerTarget ||
		!slices.Equal(strive.Mana.Symbols, []string{"{1}", "{W}"}) {
		t.Fatalf("strive effect = %#v", strive)
	}
	if len(document.Abilities[0].Sentences[0].Targets) != 0 || len(strive.Targets) != 0 {
		t.Fatal("strive cost wording declared a resolving target")
	}
	target := document.Abilities[1].Sentences[0].Targets[0]
	if target.Cardinality.Min != 0 || target.Cardinality.Max != 99 ||
		target.Selection.Kind != SelectionOpponent {
		t.Fatalf("target = %#v", target)
	}
	create := document.Abilities[1].Sentences[1].Effects[0]
	if create.Amount.DynamicKind != EffectDynamicAmountCount ||
		create.Amount.Selection == nil ||
		create.Amount.Selection.Controller != SelectionControllerTargetedPlayers {
		t.Fatalf("create amount = %#v", create.Amount)
	}
}
