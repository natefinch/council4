package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
)

// TestParseLeadingPowerToughnessPrefixTarget proves a "target N/M creature" noun
// phrase (Pendelhaven, Aegis of the Meek) parses into an exact single-creature
// target whose current power and toughness are pinned to N and M via the shared
// MatchPower/MatchToughness comparisons, with PowerToughnessPrefix recording the
// leading-"N/M" spelling for byte-exact reconstruction.
func TestParseLeadingPowerToughnessPrefixTarget(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Target 1/1 creature gets +1/+2 until end of turn.")
	if !target.Exact {
		t.Fatalf("target = %#v, want exact reconstruction", target)
	}
	sel := target.Selection
	if sel.Kind != SelectionCreature {
		t.Fatalf("selection kind = %v, want SelectionCreature", sel.Kind)
	}
	if !sel.PowerToughnessPrefix {
		t.Fatalf("selection = %#v, want PowerToughnessPrefix set", sel)
	}
	if !sel.MatchPower || sel.Power.Op != compare.Equal || sel.Power.Value != 1 {
		t.Fatalf("selection power = %#v, want exact power 1", sel.Power)
	}
	if !sel.MatchToughness || sel.Toughness.Op != compare.Equal || sel.Toughness.Value != 1 {
		t.Fatalf("selection toughness = %#v, want exact toughness 1", sel.Toughness)
	}
}

// TestParseLeadingPowerToughnessPrefixAsymmetric proves the prefix records
// distinct N and M values, so a hypothetical "target 2/3 creature" pins power to
// 2 and toughness to 3 rather than collapsing to a single value.
func TestParseLeadingPowerToughnessPrefixAsymmetric(t *testing.T) {
	t.Parallel()
	target := singleTarget(t, "Target 2/3 creature gets +1/+2 until end of turn.")
	sel := target.Selection
	if !target.Exact || !sel.PowerToughnessPrefix {
		t.Fatalf("target = %#v, want exact prefix target", target)
	}
	if sel.Power.Value != 2 || sel.Toughness.Value != 3 {
		t.Fatalf("selection = %#v, want power 2 toughness 3", sel)
	}
}
