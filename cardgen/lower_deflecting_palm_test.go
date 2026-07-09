package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDeflectingPalmRedirectShield covers Deflecting Palm's folded
// one-shot prevent-next-from-source shield with the "If damage is prevented
// this way, Deflecting Palm deals that much damage to that source's
// controller." redirect rider. The pair lowers to a single controller-scoped,
// one-shot, uncolored PreventDamage whose RedirectPreventedToSourceController
// flag carries the redirect onto the shield.
func TestLowerDeflectingPalmRedirectShield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Deflecting Palm",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}{W}",
		OracleText: "The next time a source of your choice would deal damage to you this turn, prevent that damage. If damage is prevented this way, Deflecting Palm deals that much damage to that source's controller.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("ability content = %#v, want one mode with one instruction", modes)
	}
	if len(modes[0].Targets) != 0 {
		t.Fatalf("mode targets = %#v, want no target slot", modes[0].Targets)
	}
	prevent, ok := modes[0].Sequence[0].Primitive.(game.PreventDamage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.PreventDamage", modes[0].Sequence[0].Primitive)
	}
	if !prevent.All || !prevent.OneShot || prevent.CombatOnly || prevent.BySource || prevent.Global {
		t.Fatalf("prevent = %#v, want a one-shot all-damage shield", prevent)
	}
	if prevent.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("prevent player = %#v, want the controller", prevent.Player)
	}
	if len(prevent.SourceColors) != 0 {
		t.Fatalf("prevent source colors = %#v, want none", prevent.SourceColors)
	}
	if !prevent.RedirectPreventedToSourceController {
		t.Fatalf("prevent = %#v, want RedirectPreventedToSourceController set", prevent)
	}
}
