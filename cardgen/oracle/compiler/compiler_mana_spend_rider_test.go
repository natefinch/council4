package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileManaSpendRider verifies that the parser's typed mana-spend rider
// is carried text-blind onto the compiled add-mana ability: the commander
// identity add-mana effect keeps its CommanderIdentity flag and the rider effect
// carries the typed condition, effect, and scry amount.
func TestCompileManaSpendRider(t *testing.T) {
	t.Parallel()
	source := "{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2: %#v", len(effects), effects)
	}
	if effects[0].Kind != EffectAddMana || !effects[0].Mana.CommanderIdentity {
		t.Fatalf("effect[0] = %#v, want commander-identity add-mana", effects[0])
	}
	if effects[1].Kind != EffectManaSpendRider {
		t.Fatalf("effect[1].Kind = %v, want EffectManaSpendRider", effects[1].Kind)
	}
	rider := effects[1].ManaSpendRider
	if rider == nil {
		t.Fatal("effect[1].ManaSpendRider = nil")
	}
	if rider.Condition != parser.ManaSpendCastCommanderCreatureType {
		t.Fatalf("Condition = %q, want %q", rider.Condition, parser.ManaSpendCastCommanderCreatureType)
	}
	if rider.Effect != parser.ManaSpendRiderEffectScry {
		t.Fatalf("Effect = %q, want %q", rider.Effect, parser.ManaSpendRiderEffectScry)
	}
	if rider.ScryAmount != 1 {
		t.Fatalf("ScryAmount = %d, want 1", rider.ScryAmount)
	}
}

// TestCompileManaSpendRiderNilForOtherEffects verifies that ordinary effects
// carry no rider, so the rider field is a precise signal of the recognized
// Path of Ancestry shape rather than ambient state.
func TestCompileManaSpendRiderNilForOtherEffects(t *testing.T) {
	t.Parallel()
	source := "{T}: Add one mana of any color in your commander's color identity."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, effect := range compilation.Abilities[0].Content.Effects {
		if effect.ManaSpendRider != nil {
			t.Fatalf("effect %v unexpectedly carries a rider: %#v", effect.Kind, effect.ManaSpendRider)
		}
	}
}

func TestCompileChosenTypeManaSpendRider(t *testing.T) {
	t.Parallel()
	source := "{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type, and that spell can't be countered."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2: %#v", len(effects), effects)
	}
	rider := effects[1].ManaSpendRider
	if effects[1].Kind != EffectManaSpendRider || rider == nil {
		t.Fatalf("effect[1] = %#v, want typed mana-spend rider", effects[1])
	}
	if rider.Condition != parser.ManaSpendCastChosenCreatureType ||
		rider.Effect != parser.ManaSpendRiderEffectCantBeCountered ||
		!rider.Restricted {
		t.Fatalf("rider = %#v", rider)
	}
}
