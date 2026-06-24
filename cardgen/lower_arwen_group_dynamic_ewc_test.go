package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerArwenGroupDynamicEntersWithCounters verifies that Arwen, Weaver of
// Hope's "Each other creature you control enters with a number of additional
// +1/+1 counters on it equal to Arwen's toughness." lowers to a continuous
// EntersWithCountersOthers replacement whose counter placement is dynamic
// (source toughness), the dynamic-amount sibling of the fixed "an additional
// +1/+1 counter" group form.
func TestLowerArwenGroupDynamicEntersWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Arwen, Weaver of Hope",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elf Noble",
		ManaCost:   "{3}{G}{W}",
		OracleText: "Each other creature you control enters with a number of additional +1/+1 counters on it equal to Arwen's toughness.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersWithCountersOthers {
		t.Fatalf("replacement is not a group enters-with-counters: %#v", replacement)
	}
	if replacement.EntersWithCountersRecipient == nil {
		t.Fatal("group enters-with-counters has no recipient selection")
	}
	if len(replacement.EntersWithCounters) != 1 {
		t.Fatalf("got %d counter placements, want 1", len(replacement.EntersWithCounters))
	}
	placement := replacement.EntersWithCounters[0]
	if !placement.Dynamic.Exists {
		t.Fatalf("counter placement is not dynamic: %#v", placement)
	}
	dynamic := placement.Dynamic.Val
	if dynamic.Kind != game.DynamicAmountObjectToughness {
		t.Fatalf("dynamic amount kind = %v, want object toughness", dynamic.Kind)
	}
	if len(dynamic.Object.Validate()) != 0 {
		t.Fatalf("dynamic amount object invalid: %#v", dynamic.Object)
	}
}
