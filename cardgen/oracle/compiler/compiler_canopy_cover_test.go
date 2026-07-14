package compiler

import "testing"

func TestCompileCanopyCoverStaticRules(t *testing.T) {
	t.Parallel()
	const source = "Enchanted creature can't be blocked except by creatures with flying or reach.\nEnchanted creature can't be the target of spells or abilities your opponents control."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Canopy Cover"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	first := compilation.Abilities[0].Static.Declarations[0].Rule
	if first.Kind != StaticRuleCantBeBlockedExceptBy ||
		first.Blocker.Kind != StaticBlockerRestrictionFlyingOrReach {
		t.Fatalf("block rule = %#v", first)
	}
	second := compilation.Abilities[1].Static.Declarations[0].Rule
	if second.Kind != StaticRuleCantBeTargetedByControllerOpponents ||
		second.Domain != StaticRuleDomainTarget {
		t.Fatalf("target rule = %#v", second)
	}
}
