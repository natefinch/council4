package compiler

import (
	"testing"
)

// TestCompileEnchantedPlayerFirstUpkeepTriggerPattern proves the Paradox Haze
// trigger "At the beginning of enchanted player's first upkeep each turn" compiles
// to an upkeep beginning-of-step pattern scoped to the source's enchanted player
// and gated on the first upkeep step each turn.
func TestCompileEnchantedPlayerFirstUpkeepTriggerPattern(t *testing.T) {
	t.Parallel()
	source := "At the beginning of enchanted player's first upkeep each turn, that player gets an additional upkeep step after this step."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Paradox Haze"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	trigger := compilation.Abilities[0].Trigger
	if trigger == nil {
		t.Fatal("expected a trigger")
	}

	pattern := trigger.Pattern
	if pattern.Event != TriggerEventBeginningOfStep {
		t.Errorf("event = %v, want TriggerEventBeginningOfStep", pattern.Event)
	}
	if pattern.Step != TriggerStepUpkeep {
		t.Errorf("step = %v, want TriggerStepUpkeep", pattern.Step)
	}
	if !pattern.StepPlayerIsSourceEnchantedPlayer {
		t.Error("StepPlayerIsSourceEnchantedPlayer = false, want true")
	}
	if !pattern.FirstUpkeepStepEachTurn {
		t.Error("FirstUpkeepStepEachTurn = false, want true")
	}
}

func TestCompileFirstNonUpkeepEachTurnFailsClosed(t *testing.T) {
	compilation, _ := compileSource(
		"At the beginning of your first end step each turn, you draw a card.",
		pipelineContext{},
	)
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger != nil && trigger.Pattern.Event != TriggerEventUnknown {
		t.Fatalf("trigger = %#v, want no executable first non-upkeep trigger", trigger)
	}
}

// TestCompileAdditionalUpkeepStepEffect proves the "that player gets an additional
// upkeep step after this step." effect compiles to an EffectAdditionalUpkeepStep
// carrying the AdditionalUpkeepStep flag.
func TestCompileAdditionalUpkeepStepEffect(t *testing.T) {
	t.Parallel()
	source := "At the beginning of enchanted player's first upkeep each turn, that player gets an additional upkeep step after this step."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Paradox Haze"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Kind != EffectAdditionalUpkeepStep {
		t.Errorf("kind = %v, want EffectAdditionalUpkeepStep", effect.Kind)
	}
	if !effect.AdditionalUpkeepStep {
		t.Error("AdditionalUpkeepStep = false, want true")
	}
	if !effect.Exact {
		t.Error("Exact = false, want true")
	}
}
