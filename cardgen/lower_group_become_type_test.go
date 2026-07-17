package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerGroupControlledBecomeType proves the resolving group rider a
// reanimation spell applies to its controller's creatures ("Each creature you
// control becomes a Phyrexian in addition to its other types.", Breach the
// Multiverse's closing clause) lowers into a targetless ApplyContinuous whose
// single LayerType effect adds the subtype to the controlled-creature group. The
// runtime snapshots that group's members when the spell resolves, so the grant
// is bound to a group rather than a fixed object at lowering time, adds rather
// than sets the subtype, and lasts permanently.
func TestLowerGroupControlledBecomeType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Group Become",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{5}{B}{B}",
		OracleText: "Each creature you control becomes a Phyrexian in addition to its other types.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	if len(modes[0].Targets) != 0 {
		t.Fatalf("targets = %d, want 0 (group form carries no target)", len(modes[0].Targets))
	}
	seq := modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(seq))
	}
	apply, ok := seq[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("instruction = %#v, want ApplyContinuous", seq[0].Primitive)
	}
	if apply.Object.Exists {
		t.Error("group grant must not bind a fixed object; the runtime snapshots the group")
	}
	if apply.Duration != game.DurationPermanent {
		t.Errorf("duration = %v, want DurationPermanent", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(apply.ContinuousEffects))
	}
	ce := apply.ContinuousEffects[0]
	if ce.Layer != game.LayerType {
		t.Errorf("layer = %v, want LayerType", ce.Layer)
	}
	if !slices.Equal(ce.AddSubtypes, []types.Sub{types.Phyrexian}) {
		t.Errorf("add subtypes = %v, want [Phyrexian]", ce.AddSubtypes)
	}
	if len(ce.AddTypes) != 0 || len(ce.AddColors) != 0 {
		t.Errorf("add types = %v, add colors = %v, want none", ce.AddTypes, ce.AddColors)
	}
	if ce.AffectedObjectID != 0 {
		t.Error("group member must be snapshotted at resolution, not fixed at lowering time")
	}
	assertControlledCreatureGroup(t, ce.Group)
}

// TestLowerGroupControlledBecomeTypeAfterReturn proves the group color-and-type
// rider fuses onto a reanimation put: the returned card enters through
// PutOnBattlefield, then a targetless ApplyContinuous adds the color at LayerColor
// and the subtype at LayerType to every creature its controller controls at
// resolution.
func TestLowerGroupControlledBecomeTypeAfterReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Return Then Group Become",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{5}{B}{B}",
		OracleText: "Put target creature card from a graveyard onto the battlefield under your control. " +
			"Then each creature you control becomes a black Zombie in addition to its other colors and types.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (the returned card)", len(mode.Targets))
	}
	seq := mode.Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2 (put + group become)", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.PutOnBattlefield); !ok {
		t.Fatalf("first instruction = %#v, want PutOnBattlefield", seq[0].Primitive)
	}
	apply, ok := seq[1].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("second instruction = %#v, want ApplyContinuous", seq[1].Primitive)
	}
	if apply.Object.Exists {
		t.Error("group grant must not bind a fixed object")
	}
	if apply.Duration != game.DurationPermanent {
		t.Errorf("duration = %v, want DurationPermanent", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %d, want 2 (type + color)", len(apply.ContinuousEffects))
	}
	typeEffect := apply.ContinuousEffects[0]
	if typeEffect.Layer != game.LayerType || !slices.Equal(typeEffect.AddSubtypes, []types.Sub{types.Zombie}) {
		t.Errorf("type effect = %+v, want LayerType adding [Zombie]", typeEffect)
	}
	colorEffect := apply.ContinuousEffects[1]
	if colorEffect.Layer != game.LayerColor || !slices.Equal(colorEffect.AddColors, []color.Color{color.Black}) {
		t.Errorf("color effect = %+v, want LayerColor adding [Black]", colorEffect)
	}
	assertControlledCreatureGroup(t, typeEffect.Group)
	assertControlledCreatureGroup(t, colorEffect.Group)
}

// assertControlledCreatureGroup asserts the group names every creature its
// controller controls: a battlefield-domain group filtered to creatures the
// spell's controller controls.
func assertControlledCreatureGroup(t *testing.T, group game.GroupReference) {
	t.Helper()
	if group.Empty() {
		t.Fatal("continuous effect carries no group")
	}
	if group.Domain() != game.GroupDomainBattlefield {
		t.Errorf("group domain = %v, want GroupDomainBattlefield", group.Domain())
	}
	selection := group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Errorf("group controller = %v, want ControllerYou", selection.Controller)
	}
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Errorf("group required types = %v, want [Creature]", selection.RequiredTypes)
	}
}
