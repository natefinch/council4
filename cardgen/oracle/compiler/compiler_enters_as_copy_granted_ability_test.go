package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileEntersAsCopyGrantedAbilityRider proves an enters-as-copy effect
// carrying an "except it has \"<quoted ability>\"" rider compiles to an
// EffectEnterAsCopy that records the granted-ability marker and binds the parsed
// quoted ability, so lowering can attach it to the copy's copiable values.
func TestCompileEntersAsCopyGrantedAbilityRider(t *testing.T) {
	t.Parallel()
	text := "You may have this enchantment enter as a copy of an enchantment you control, " +
		"except it has \"At the beginning of your upkeep, you may exile this enchantment. " +
		"If you do, return it to the battlefield under its owner's control.\""
	compilation, diagnostics := compileSource(text, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectEnterAsCopy {
		t.Fatalf("effects = %#v, want single enters-as-copy", effects)
	}
	effect := effects[0]
	if !effect.EntersAsCopyGrantedAbilityRider {
		t.Fatal("EntersAsCopyGrantedAbilityRider = false, want true")
	}
	if effect.EntersAsCopyGrantedAbility == nil {
		t.Fatal("EntersAsCopyGrantedAbility is nil, want the parsed quoted upkeep ability")
	}
	document, docDiagnostics := effect.EntersAsCopyGrantedAbility.Inner()
	if len(docDiagnostics) != 0 {
		t.Fatalf("granted ability inner diagnostics = %#v", docDiagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("granted ability document = %#v, want one ability", document)
	}
	if document.Abilities[0].Kind != parser.AbilityTriggered {
		t.Fatalf("granted ability kind = %v, want triggered", document.Abilities[0].Kind)
	}
}
