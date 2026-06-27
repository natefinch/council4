package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// The condition-first source-conditional self-anthem "As long as <source-state>,
// it gets +X/+Y and/or has <keyword(s)>" names the source permanent through the
// pronoun "it" rather than the trailing "This creature ... as long as it's
// equipped" wording. Both forms grant the source the same continuous effect, so
// the pronoun subject must resolve to the source and compile without a static
// declaration group diagnostic.

func TestCompileSourcePronounConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	source := "As long as this creature is equipped, it has flying."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource {
		t.Fatalf("declaration.Group = %#v", declaration.Group)
	}
	if declaration.Condition == nil ||
		declaration.Continuous == nil ||
		declaration.Continuous.Operation != StaticContinuousGrantKeywords ||
		len(declaration.Continuous.Keywords) != 1 ||
		declaration.Continuous.Keywords[0].Kind != parser.KeywordFlying {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestCompileSourcePronounConditionalPowerToughnessGrant(t *testing.T) {
	t.Parallel()
	source := "As long as this creature is equipped, it gets +1/+1."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Condition == nil ||
		declaration.Continuous == nil ||
		declaration.Continuous.Operation != StaticContinuousModifyPowerToughness ||
		declaration.Continuous.PowerDelta.Value != 1 ||
		declaration.Continuous.ToughnessDelta.Value != 1 {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestCompileSourcePronounConditionalComposedGrant(t *testing.T) {
	t.Parallel()
	source := "As long as this creature is equipped, it gets +1/+1 and has flying."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 2 {
		t.Fatalf("ability = %#v", ability)
	}
	for _, declaration := range ability.Static.Declarations {
		if declaration.Group.Domain != StaticGroupSource || declaration.Condition == nil || declaration.Continuous == nil {
			t.Fatalf("declaration = %#v", declaration)
		}
	}
}

func TestCompileSourcePronounConditionalEnchantedGrant(t *testing.T) {
	t.Parallel()
	source := "As long as this creature is enchanted, it has flying."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Condition == nil ||
		declaration.Continuous == nil ||
		declaration.Continuous.Operation != StaticContinuousGrantKeywords {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestCompileSourcePronounConditionalCounterMultiKeywordGrant(t *testing.T) {
	t.Parallel()
	source := "As long as this creature has four or more +1/+1 counters on it, it has flying and vigilance."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Condition == nil ||
		declaration.Continuous == nil ||
		declaration.Continuous.Operation != StaticContinuousGrantKeywords ||
		len(declaration.Continuous.Keywords) != 2 {
		t.Fatalf("declaration = %#v", declaration)
	}
}

// The source pronoun path must not hijack the existing attached-object pronoun
// form. "As long as equipped creature is <state>, it ..." binds the attached
// object, not the source, so its declaration must keep the attached-object
// domain rather than resolving "it" to the source permanent.
func TestCompileAttachedPronounUnaffectedBySourcePronounPath(t *testing.T) {
	t.Parallel()
	source := "As long as equipped creature is tapped, it gets +1/+1."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupAttachedObject {
		t.Fatalf("declaration.Group = %#v", declaration.Group)
	}
}
