package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func groupModifyPTContinuous(t *testing.T, oracleText string) game.ContinuousEffect {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Group Modify",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	primitive, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if primitive.Object.Exists || primitive.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("primitive = %+v, want unanchored group effect until end of turn", primitive)
	}
	if len(primitive.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(primitive.ContinuousEffects))
	}
	effect := primitive.ContinuousEffects[0]
	if effect.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", effect.Layer)
	}
	return effect
}

func TestLowerGroupModifyPTAllCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "All creatures get -1/-1 until end of turn.")
	if effect.PowerDelta != -1 || effect.ToughnessDelta != -1 {
		t.Fatalf("delta = %d/%d, want -1/-1", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerAny ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want every creature on the battlefield", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("all creatures must not exclude the source")
	}
}

func TestLowerGroupModifyPTEachCreature(t *testing.T) {
	t.Parallel()
	// "Each creature gets ..." is the singular distributive wording for the same
	// every-creature group as "All creatures get ...".
	effect := groupModifyPTContinuous(t, "Each creature gets -1/-1 until end of turn.")
	if effect.PowerDelta != -1 || effect.ToughnessDelta != -1 {
		t.Fatalf("delta = %d/%d, want -1/-1", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerAny ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want every creature on the battlefield", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("each creature must not exclude the source")
	}
}

func TestLowerGroupModifyPTAllOtherCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "All other creatures get -2/-2 until end of turn.")
	if effect.PowerDelta != -2 || effect.ToughnessDelta != -2 {
		t.Fatalf("delta = %d/%d, want -2/-2", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerAny ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want every creature (any controller)", selection)
	}
	exclude, excludes := effect.Group.Exclusion()
	if !excludes || exclude != game.SourcePermanentReference() {
		t.Fatalf("exclusion = %v/%v, want source permanent excluded", exclude, excludes)
	}
}

func TestLowerGroupModifyPTOtherControlledCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Other creatures you control get +1/+1 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 1 {
		t.Fatalf("delta = %d/%d, want +1/+1", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
	exclude, excludes := effect.Group.Exclusion()
	if !excludes || exclude != game.SourcePermanentReference() {
		t.Fatalf("exclusion = %v/%v, want source permanent excluded", exclude, excludes)
	}
}

func TestLowerGroupModifyPTControlledCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Creatures you control get +1/+0 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +1/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("creatures you control must not exclude the source")
	}
}

func TestLowerGroupModifyPTColorFiltered(t *testing.T) {
	t.Parallel()
	tests := []struct {
		oracleText string
		controller game.ControllerRelation
		wantColor  color.Color
	}{
		{"White creatures you control get +2/+2 until end of turn.", game.ControllerYou, color.White},
		{"Black creatures get +2/+0 until end of turn.", game.ControllerAny, color.Black},
	}
	for _, test := range tests {
		effect := groupModifyPTContinuous(t, test.oracleText)
		selection := effect.Group.Selection()
		if selection.Controller != test.controller ||
			!slices.Equal(selection.ColorsAny, []color.Color{test.wantColor}) {
			t.Fatalf("selection = %#v", selection)
		}
	}
}

func TestLowerGroupModifyPTControlledCreatureSubtype(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Dragons you control get +1/+0 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +1/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.SubtypesAny) != 1 ||
		selection.SubtypesAny[0] != types.Dragon {
		t.Fatalf("selection = %+v, want Dragons you control", selection)
	}
}

func TestLowerGroupModifyPTBattlefieldCreatureSubtype(t *testing.T) {
	t.Parallel()
	// "Goblin creatures get ..." names every Goblin on the battlefield, the
	// unprefixed sibling of "All Goblin creatures get ...".
	effect := groupModifyPTContinuous(t, "Goblin creatures get +1/+1 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 1 {
		t.Fatalf("delta = %d/%d, want +1/+1", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerAny ||
		len(selection.SubtypesAny) != 1 ||
		selection.SubtypesAny[0] != types.Goblin {
		t.Fatalf("selection = %+v, want every Goblin on the battlefield", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("battlefield Goblins must not exclude the source")
	}
}

func TestLowerGroupModifyPTBattlefieldNonSubtype(t *testing.T) {
	t.Parallel()
	// "Non-Elf creatures get ..." names every creature on the battlefield that
	// does not carry the Elf subtype.
	effect := groupModifyPTContinuous(t, "Non-Elf creatures get -2/-2 until end of turn.")
	if effect.PowerDelta != -2 || effect.ToughnessDelta != -2 {
		t.Fatalf("delta = %d/%d, want -2/-2", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerAny ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature ||
		selection.ExcludedSubtype != types.Elf {
		t.Fatalf("selection = %+v, want every non-Elf creature on the battlefield", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("non-Elf creatures must not exclude the source")
	}
}

func TestLowerGroupModifyPTEachControlledCreature(t *testing.T) {
	t.Parallel()
	// "Each creature you control gets ..." is the singular distributive wording
	// for the same group as "Creatures you control get ...".
	effect := groupModifyPTContinuous(t, "Each creature you control gets +1/+0 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +1/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
	if _, excludes := effect.Group.Exclusion(); excludes {
		t.Fatal("each creature you control must not exclude the source")
	}
}

func TestLowerGroupModifyPTEachOtherControlledCreature(t *testing.T) {
	t.Parallel()
	// "Each other creature you control gets ..." is the singular distributive
	// wording for the same source-excluding group as "Other creatures you
	// control get ...".
	effect := groupModifyPTContinuous(t, "Each other creature you control gets +1/+0 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +1/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
	exclude, excludes := effect.Group.Exclusion()
	if !excludes || exclude != game.SourcePermanentReference() {
		t.Fatalf("exclusion = %v/%v, want source permanent excluded", exclude, excludes)
	}
}

func TestLowerGroupModifyPTAttackingCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Attacking creatures get +2/+0 until end of turn.")
	if effect.PowerDelta != 2 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +2/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerAny ||
		selection.CombatState != game.CombatStateAttacking ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want attacking creatures", selection)
	}
}

func TestLowerGroupModifyPTBlockingCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Blocking creatures get +0/+3 until end of turn.")
	if effect.PowerDelta != 0 || effect.ToughnessDelta != 3 {
		t.Fatalf("delta = %d/%d, want +0/+3", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if selection.Controller != game.ControllerAny ||
		selection.CombatState != game.CombatStateBlocking ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want blocking creatures", selection)
	}
}

func groupModifyPTUnsupported(t *testing.T, oracleText string) {
	t.Helper()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Group Reject",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	if len(diagnostics) == 0 {
		t.Fatalf("expected fail-closed diagnostic for %q", oracleText)
	}
}

func TestLowerGroupModifyPTFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []string{
		// Variable / dynamic group amount.
		"All creatures get -X/-X until end of turn.",
		"Each creature gets -X/-X until end of turn.",
		"Creatures you control get +X/+X until end of turn.",
		// Excluded-color groups are not yet representable.
		"Nongreen creatures you control get +1/+0 until end of turn.",
		// Rider beyond the bare power/toughness change.
		"All creatures get -1/-1 until end of turn and can't block this turn.",
		// Conditional group buff.
		"If you control a Mountain, creatures you control get +1/+0 until end of turn.",
	}
	for _, oracleText := range cases {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			groupModifyPTUnsupported(t, oracleText)
		})
	}
}

func TestLowerDoubleGroupPowerToughness(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Double the power and toughness of each creature you control until end of turn.")
	if !effect.DoublePower || !effect.DoubleToughness {
		t.Fatalf("doublePower=%v doubleToughness=%v, want both true", effect.DoublePower, effect.DoubleToughness)
	}
	if effect.PowerDelta != 0 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want 0/0 (doubling carries no fixed delta)", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creatures you control", selection)
	}
}

func TestLowerDoubleGroupPowerOnly(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Double the power of each creature you control until end of turn.")
	if !effect.DoublePower || effect.DoubleToughness {
		t.Fatalf("doublePower=%v doubleToughness=%v, want power only", effect.DoublePower, effect.DoubleToughness)
	}
}

// TestLowerGroupModifyPTOtherAttackingCreatures covers the Battle cry group:
// each other attacking creature gets +1/+0, excluding the source.
func TestLowerGroupModifyPTOtherAttackingCreatures(t *testing.T) {
	t.Parallel()
	effect := groupModifyPTContinuous(t, "Each other attacking creature gets +1/+0 until end of turn.")
	if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
		t.Fatalf("delta = %d/%d, want +1/+0", effect.PowerDelta, effect.ToughnessDelta)
	}
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.CombatState != game.CombatStateAttacking ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want attacking creatures", selection)
	}
	exclude, excludes := effect.Group.Exclusion()
	if !excludes || exclude != game.SourcePermanentReference() {
		t.Fatalf("exclusion = %v/%v, want source permanent excluded", exclude, excludes)
	}
}
