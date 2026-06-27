package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// TestLowerPreventNextDamageFromColoredSource covers the one-shot shield "The
// next time a <color> source of your choice would deal damage to you this turn,
// prevent that damage." (Circle of Protection, Rune of Protection), which lowers
// to a controller-scoped, one-shot, color-filtered PreventDamage that prevents
// all of the next matching damage event.
func TestLowerPreventNextDamageFromColoredSource(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Circle of Protection: White",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	modes := face.ActivatedAbilities[0].Content.Modes
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
	if want := []color.Color{color.White}; len(prevent.SourceColors) != 1 || prevent.SourceColors[0] != want[0] {
		t.Fatalf("prevent source colors = %#v, want %#v", prevent.SourceColors, want)
	}
}

// TestLowerPreventNextDamageFromAnySource covers the unfiltered form ("a source
// of your choice"), which lowers to the same one-shot shield with no source
// color filter.
func TestLowerPreventNextDamageFromAnySource(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Pentagram of the Ages",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{4}, {T}: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	modes := face.ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("ability content = %#v, want one mode with one instruction", modes)
	}
	prevent, ok := modes[0].Sequence[0].Primitive.(game.PreventDamage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.PreventDamage", modes[0].Sequence[0].Primitive)
	}
	if !prevent.All || !prevent.OneShot {
		t.Fatalf("prevent = %#v, want a one-shot all-damage shield", prevent)
	}
	if len(prevent.SourceColors) != 0 {
		t.Fatalf("prevent source colors = %#v, want none", prevent.SourceColors)
	}
}
