package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// referencedDynamicModifyPT lowers a card whose single activated or triggered
// ability pumps the source ("This creature gets …") or the triggering permanent
// ("it gets …") by a dynamic until-end-of-turn amount, returning the ModifyPT
// primitive. activated selects which ability slot to read.
func referencedDynamicModifyPT(t *testing.T, typeLine, oracleText string, activated bool) game.ModifyPT {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Referenced Dynamic Pump",
		Layout:     "normal",
		TypeLine:   typeLine,
		OracleText: oracleText,
	})
	var content game.AbilityContent
	if activated {
		content = face.ActivatedAbilities[0].Content
	} else {
		content = face.TriggeredAbilities[0].Content
	}
	mode := content.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", modify.Duration)
	}
	return modify
}

func TestLowerSourceDynamicWhereXSelfPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Creature — Dragon",
		"{1}{R}: This creature gets +X/+0 until end of turn, where X is the number of artifacts you control.",
		true)
	if modify.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %+v, want source permanent reference", modify.Object)
	}
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists || power.Val.Kind != game.DynamicAmountCountSelector || power.Val.Multiplier != 1 {
		t.Fatalf("power delta = %+v, want count-selector multiplier 1", modify.PowerDelta)
	}
	assertControlledTypeGroup(t, "power", power.Val.Group, types.Artifact)
	if modify.ToughnessDelta.IsDynamic() || modify.ToughnessDelta.Value() != 0 {
		t.Fatalf("toughness delta = %+v, want fixed 0", modify.ToughnessDelta)
	}
}

func TestLowerSourceDynamicForEachSelfPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Creature — Human Knight",
		"{2}{G}: This creature gets +1/+1 until end of turn for each enchantment you control.",
		true)
	if modify.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %+v, want source permanent reference", modify.Object)
	}
	for _, side := range []struct {
		name     string
		quantity game.Quantity
	}{{"power", modify.PowerDelta}, {"toughness", modify.ToughnessDelta}} {
		dynamic := side.quantity.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector || dynamic.Val.Multiplier != 1 {
			t.Fatalf("%s delta = %+v, want count-selector multiplier 1", side.name, side.quantity)
		}
		assertControlledTypeGroup(t, side.name, dynamic.Val.Group, types.Enchantment)
	}
}

func TestLowerEventPermanentDynamicWhereXPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Enchantment",
		"Whenever a creature you control attacks alone, it gets +X/+X until end of turn, where X is the number of creatures you control.",
		false)
	if modify.Object != game.EventPermanentReference() {
		t.Fatalf("object = %+v, want event permanent reference", modify.Object)
	}
	for _, side := range []struct {
		name     string
		quantity game.Quantity
	}{{"power", modify.PowerDelta}, {"toughness", modify.ToughnessDelta}} {
		dynamic := side.quantity.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector {
			t.Fatalf("%s delta = %+v, want count-selector", side.name, side.quantity)
		}
		assertControlledTypeGroup(t, side.name, dynamic.Val.Group, types.Creature)
	}
}

// TestLowerSharedAnimosityAttacksPump lowers Shared Animosity's attack-trigger
// self-pump scaled by the number of other attacking creatures that share a
// creature type with the triggering attacker. The pump addresses the triggering
// permanent and reads a shared-creature-type count scoped to attacking creatures.
func TestLowerSharedAnimosityAttacksPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Enchantment",
		"Whenever a creature you control attacks, it gets +1/+0 until end of turn for each other attacking creature that shares a creature type with it.",
		false)
	if modify.Object != game.EventPermanentReference() {
		t.Fatalf("object = %+v, want event permanent reference", modify.Object)
	}
	power := modify.PowerDelta.DynamicAmount()
	if !power.Exists || power.Val.Kind != game.DynamicAmountSharedCreatureTypeCountInGroup || power.Val.Multiplier != 1 {
		t.Fatalf("power delta = %+v, want shared-creature-type count multiplier 1", modify.PowerDelta)
	}
	if state := power.Val.Group.Selection().CombatState; state != game.CombatStateAttacking {
		t.Fatalf("power group combat state = %v, want attacking", state)
	}
	if modify.ToughnessDelta.IsDynamic() || modify.ToughnessDelta.Value() != 0 {
		t.Fatalf("toughness delta = %+v, want fixed 0", modify.ToughnessDelta)
	}
}

// TestLowerSourceDynamicSourcePowerSelfPump lowers a self-pump scaled by the
// source's own power ("This creature gets +X/+X until end of turn, where X is its
// power."). The pump addresses the source and reads the source's power, which the
// runtime snapshots at resolution.
func TestLowerSourceDynamicSourcePowerSelfPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Creature — Elemental",
		"{2}{G}: This creature gets +X/+X until end of turn, where X is its power.",
		true)
	if modify.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %+v, want source permanent reference", modify.Object)
	}
	for _, side := range []struct {
		name     string
		quantity game.Quantity
	}{{"power", modify.PowerDelta}, {"toughness", modify.ToughnessDelta}} {
		dynamic := side.quantity.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectPower ||
			dynamic.Val.Multiplier != 1 ||
			dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%s delta = %+v, want source object-power multiplier 1", side.name, side.quantity)
		}
	}
}

// TestLowerSourcePowerPumpKeywordRiderRejected keeps a source-power pump that also
// grants a keyword fail closed: the source-power lowering models only the bare
// power/toughness change, so a "… and gains …" rider must not be silently
// dropped.
func TestLowerSourcePowerPumpKeywordRiderRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Power Pump Rider",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +X/+X and gains trample until end of turn, where X is its power.",
	})
	if len(diagnostics) == 0 {
		t.Fatalf("diagnostics = %+v, want the keyword rider to fail closed", diagnostics)
	}
}

// assertControlledTypeGroup checks that group counts battlefield permanents you
// control of exactly the given single required card type.
func assertControlledTypeGroup(t *testing.T, label string, group game.GroupReference, want types.Card) {
	t.Helper()
	if group.Domain() != game.GroupDomainBattlefield {
		t.Fatalf("%s group domain = %v, want battlefield", label, group.Domain())
	}
	selection := group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != want {
		t.Fatalf("%s group required types = %+v, want [%v]", label, selection.RequiredTypes, want)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("%s group controller = %v, want you", label, selection.Controller)
	}
}

func hasReferencedPTDiagnostic(diagnostics []shared.Diagnostic, summary string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Summary == summary {
			return true
		}
	}
	return false
}
