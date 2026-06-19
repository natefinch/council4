package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// lifeRiderDynamic extracts the dynamic amount of a GainLife/LoseLife primitive,
// failing the test when the quantity is not the expected object-characteristic
// dynamic amount.
func lifeRiderDynamic(t *testing.T, amount game.Quantity) game.DynamicAmount {
	t.Helper()
	dyn := amount.DynamicAmount()
	if !dyn.Exists {
		t.Fatalf("amount %+v is not dynamic", amount)
	}
	return dyn.Val
}

// TestLowerSwordsToPlowsharesLifeRider verifies the marquee shape: "Exile target
// creature. Its controller gains life equal to its power." lowers to an exile
// that publishes the exiled creature under a linked key, followed by a GainLife
// whose recipient is the exiled creature's controller and whose amount is that
// creature's (last-known) power.
func TestLowerSwordsToPlowsharesLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Swords",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature. Its controller gains life equal to its power.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first primitive = %+v, want exile of target 0", mode.Sequence[0].Primitive)
	}
	if exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %+v, want a published ExileLinkedKey", exile)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	wantPlayer := game.ObjectControllerReference(game.TargetPermanentReference(0))
	if gain.Player != wantPlayer {
		t.Fatalf("gain player = %+v, want %+v", gain.Player, wantPlayer)
	}
	dyn := lifeRiderDynamic(t, gain.Amount)
	if dyn.Kind != game.DynamicAmountObjectPower || dyn.Multiplier != 1 {
		t.Fatalf("dynamic = %+v, want ObjectPower multiplier 1", dyn)
	}
	if dyn.Object != game.LinkedObjectReference(string(exile.ExileLinkedKey)) {
		t.Fatalf("dynamic object = %+v, want linked object %q", dyn.Object, exile.ExileLinkedKey)
	}
}

// TestLowerChastiseLifeRider verifies the controller-recipient, target-bound
// variant: "Destroy target attacking creature. You gain life equal to its
// power." gains life equal to the destroyed creature's power for the spell's
// controller, with the amount read directly from the target permanent.
func TestLowerChastiseLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chastise",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target attacking creature. You gain life equal to its power.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if destroy, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok || destroy.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first primitive = %+v, want destroy of target 0", mode.Sequence[0].Primitive)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	if gain.Player != game.ControllerReference() {
		t.Fatalf("gain player = %+v, want controller", gain.Player)
	}
	dyn := lifeRiderDynamic(t, gain.Amount)
	if dyn.Kind != game.DynamicAmountObjectPower || dyn.Object != game.TargetPermanentReference(0) {
		t.Fatalf("dynamic = %+v, want ObjectPower of target 0", dyn)
	}
}

// TestLowerToughnessLifeRider verifies the toughness sibling of the shape (the
// Avenger en-Dal characteristic): "Exile target creature. Its controller gains
// life equal to its toughness." reads the exiled creature's toughness.
func TestLowerToughnessLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Toughness Rider",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature. Its controller gains life equal to its toughness.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	dyn := lifeRiderDynamic(t, gain.Amount)
	if dyn.Kind != game.DynamicAmountObjectToughness {
		t.Fatalf("dynamic = %+v, want ObjectToughness", dyn)
	}
}

// TestLifeRiderFailsClosed verifies the rider stays fail-closed for amount
// characteristics and recipients it does not model: a mana-value amount, a
// fixed-amount rider (handled elsewhere), and a lone life clause with no prior
// subject to bind "its" to.
func TestLifeRiderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// Mana value is not a modeled characteristic for this shape.
		"Exile target creature. Its controller gains life equal to its mana value.",
		// No earlier clause defines the antecedent for "its power".
		"You gain life equal to its power.",
		// "its power" amount but the recipient is an unmodeled targeted player.
		"Exile target creature. Target player gains life equal to its power.",
	}
	for _, text := range rejected {
		faces, _ := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: text,
		})
		for _, face := range faces {
			if face.SpellAbility.Exists {
				t.Errorf("OracleText %q lowered a spell ability, want fail closed", text)
			}
		}
	}
}
