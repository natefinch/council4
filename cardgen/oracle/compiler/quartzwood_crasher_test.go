package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileQuartzwoodCombatDamageBatchTrigger proves the compiler carries the
// reusable combat-damage batch semantics onto Quartzwood Crasher's trigger
// pattern: OneOrMorePerDamagedPlayer coalesces simultaneous combat damage to the
// same player, the source filter requires trample, and the token is sized by the
// TriggeringEventTotalCombatDamage dynamic amount.
func TestCompileQuartzwoodCombatDamageBatchTrigger(t *testing.T) {
	t.Parallel()
	source := "Trample\n" +
		"Whenever one or more creatures you control with trample deal combat damage to a player, " +
		"create an X/X green Dinosaur Beast creature token with trample, " +
		"where X is the amount of damage those creatures dealt to that player."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Quartzwood Crasher"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	var triggered *CompiledAbility
	for i := range compilation.Abilities {
		if compilation.Abilities[i].Trigger != nil {
			triggered = &compilation.Abilities[i]
			break
		}
	}
	if triggered == nil {
		t.Fatal("no triggered ability compiled")
	}

	pattern := triggered.Trigger.Pattern
	if pattern.Event != TriggerEventDamageDealt {
		t.Errorf("event = %v, want TriggerEventDamageDealt", pattern.Event)
	}
	if !pattern.OneOrMore {
		t.Error("OneOrMore = false, want true")
	}
	if !pattern.OneOrMorePerDamagedPlayer {
		t.Error("OneOrMorePerDamagedPlayer = false, want true (per-player combat-damage coalescing)")
	}
	if pattern.CombatQualifier != TriggerCombatDamage {
		t.Errorf("combat qualifier = %v, want TriggerCombatDamage", pattern.CombatQualifier)
	}
	if pattern.DamageRecipient != TriggerDamageRecipientPlayer {
		t.Errorf("damage recipient = %v, want TriggerDamageRecipientPlayer", pattern.DamageRecipient)
	}
	if pattern.DamageRecipientIsSource {
		t.Error("DamageRecipientIsSource = true, want false (an independent player)")
	}
	if pattern.DamageSourceSelection.Keyword != parser.KeywordTrample {
		t.Errorf("damage source keyword = %v, want KeywordTrample", pattern.DamageSourceSelection.Keyword)
	}

	effect := quartzwoodCompiledCreate(t, triggered.Content.Effects)
	if effect.Amount.DynamicKind != DynamicAmountTriggeringEventTotalCombatDamage {
		t.Errorf("amount dynamic kind = %v, want DynamicAmountTriggeringEventTotalCombatDamage", effect.Amount.DynamicKind)
	}
	if !effect.TokenPTVariableX {
		t.Error("TokenPTVariableX = false, want true (X/X token)")
	}
}

// quartzwoodCompiledCreate returns the sole EffectCreate among the compiled
// effects.
func quartzwoodCompiledCreate(t *testing.T, effects []CompiledEffect) CompiledEffect {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == EffectCreate {
			return effect
		}
	}
	t.Fatal("no EffectCreate compiled")
	return CompiledEffect{}
}
