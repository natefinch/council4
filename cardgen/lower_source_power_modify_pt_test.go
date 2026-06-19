package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// spellSourcePowerModifyPT lowers an instant or sorcery whose single targeted
// power/toughness pump scales by a permanent's power ("… where X is its power.")
// and returns the ModifyPT primitive.
func spellSourcePowerModifyPT(t *testing.T, oracleText string) game.ModifyPT {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Source Power Pump",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %+v, want one single-creature target", mode.Targets)
	}
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %+v, want target permanent reference 0", modify.Object)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", modify.Duration)
	}
	return modify
}

// TestLowerTargetSourcePowerPumpTargetItsPower covers the "where X is its power"
// target pump (Rush of Blood), which reads the target's own power.
func TestLowerTargetSourcePowerPumpTargetItsPower(t *testing.T) {
	t.Parallel()
	modify := spellSourcePowerModifyPT(t,
		"Target creature gets +X/+0 until end of turn, where X is its power.")
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists || power.Val.Kind != game.DynamicAmountObjectPower ||
		power.Val.Multiplier != 1 || power.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("power delta = %+v, want target object-power multiplier 1", modify.PowerDelta)
	}
	if modify.ToughnessDelta.IsDynamic() || modify.ToughnessDelta.Value() != 0 {
		t.Fatalf("toughness delta = %+v, want fixed 0", modify.ToughnessDelta)
	}
}

// TestLowerActivatedSourcePowerPumpThisCreaturePower covers an activated ability
// whose target pump scales by the source's power ("where X is this creature's
// power", e.g. Auriok Bladewarden): the pumped object is the target while the
// power amount reads the source.
func TestLowerActivatedSourcePowerPumpThisCreaturePower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bladewarden",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "{T}: Target creature gets +X/+X until end of turn, where X is this creature's power.",
	})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %+v, want target permanent reference 0", modify.Object)
	}
	for _, side := range []struct {
		name     string
		quantity game.Quantity
	}{{"power", modify.PowerDelta}, {"toughness", modify.ToughnessDelta}} {
		dynamic := side.quantity.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectPower ||
			dynamic.Val.Multiplier != 1 || dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%s delta = %+v, want source object-power multiplier 1", side.name, side.quantity)
		}
	}
}

// TestLowerTriggeredSourcePowerPumpNamePower covers a triggered ability whose
// target pump scales by the source named explicitly ("where X is <Name>'s
// power", e.g. Syr Faren): the power amount reads the source.
func TestLowerTriggeredSourcePowerPumpNamePower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Hengehammer",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "Whenever Hengehammer attacks, another target attacking creature gets +X/+X until end of turn, where X is Hengehammer's power.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %+v, want target permanent reference 0", modify.Object)
	}
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists || power.Val.Kind != game.DynamicAmountObjectPower ||
		power.Val.Object != game.SourcePermanentReference() {
		t.Fatalf("power delta = %+v, want source object-power", modify.PowerDelta)
	}
}

// TestLowerSourcePowerPumpNonCreatureTargetRejected keeps a source-power pump
// fail closed when the target is not a creature, since the bounded form supports
// only a single creature target.
func TestLowerSourcePowerPumpNonCreatureTargetRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Artifact Power Pump",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target artifact creature gets +X/+0 until end of turn, where X is its power.",
	})
	if !hasReferencedPTDiagnostic(diagnostics, "unsupported power/toughness spell") {
		t.Fatalf("diagnostics = %+v, want unsupported power/toughness spell", diagnostics)
	}
}
