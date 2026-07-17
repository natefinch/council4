package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestCompileMultiTokenSharedVariableX proves the compiler carries a shared-count
// multi-token create through to a CompiledEffect whose AdditionalTokens holds the
// later token spec, with both the leading spec and the additional spec preserving
// the spell's variable X count. This confirms the multi-token specs survive
// compilation (via compileEffects over AdditionalTokens) with the shared X intact
// so the lowering can emit one CreateToken per spec at the same quantity.
func TestCompileMultiTokenSharedVariableX(t *testing.T) {
	t.Parallel()
	source := `When this creature enters, create X 1/1 white Halfling creature tokens and X Food tokens. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")`
	document, diagnostics := parser.Parse(source, parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) == 0 || effects[0].Kind != EffectCreate {
		t.Fatalf("compiled effects = %#v, want a leading create", effects)
	}
	create := effects[0]
	if !create.Amount.VariableX {
		t.Errorf("leading spec amount = %+v, want VariableX", create.Amount)
	}
	if len(create.AdditionalTokens) != 1 {
		t.Fatalf("AdditionalTokens = %d, want 1", len(create.AdditionalTokens))
	}
	food := create.AdditionalTokens[0]
	if food.Kind != EffectCreate {
		t.Errorf("additional token kind = %v, want EffectCreate", food.Kind)
	}
	if !food.Amount.VariableX {
		t.Errorf("Food spec amount = %+v, want VariableX", food.Amount)
	}
}

func TestCompileCrewPowerContributionStatic(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"This creature crews Vehicles as though its power were 2 greater.",
		parser.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("compiled ability = %#v, want one static declaration", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationCrewPowerContribution ||
		declaration.CrewPowerContribution == nil ||
		declaration.CrewPowerContribution.Bonus != 2 {
		t.Fatalf("compiled declaration = %#v", declaration)
	}
}

func TestCompileCrewPowerContributionIsTextAndPositionBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"This creature crews Vehicles as though its power were 2 greater.",
		parser.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	ability := &document.Abilities[0]
	ability.Text = "unrelated metadata"
	ability.Tokens = nil
	ability.Span = shared.Span{}
	ability.StaticDeclarations[0].Span = shared.Span{}
	ability.StaticDeclarations[0].OperationSpan = shared.Span{}
	ability.StaticDeclarations[0].Subject.Span = shared.Span{}

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	declaration := compilation.Abilities[0].Static.Declarations[0]
	if declaration.CrewPowerContribution == nil ||
		declaration.CrewPowerContribution.Bonus != 2 {
		t.Fatalf("compiled declaration = %#v", declaration)
	}
}
