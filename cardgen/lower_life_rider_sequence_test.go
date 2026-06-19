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

// TestLowerFeedTheSwarmManaValueLifeRider verifies the mana-value sibling of the
// shape (the Feed the Swarm characteristic): "Destroy target permanent. You lose
// life equal to its mana value." lowers to a destroy of the target permanent
// followed by a LoseLife for the controller, read from the destroyed permanent's
// (last-known) mana value.
func TestLowerFeedTheSwarmManaValueLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Feed the Swarm",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature or enchantment. You lose life equal to its mana value.",
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
	lose, ok := mode.Sequence[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.LoseLife", mode.Sequence[1].Primitive)
	}
	if lose.Player != game.ControllerReference() {
		t.Fatalf("lose player = %+v, want controller", lose.Player)
	}
	dyn := lifeRiderDynamic(t, lose.Amount)
	if dyn.Kind != game.DynamicAmountObjectManaValue || dyn.Object != game.TargetPermanentReference(0) {
		t.Fatalf("dynamic = %+v, want ObjectManaValue of target 0", dyn)
	}
}

// TestLowerDivineOfferingManaValueLifeRider verifies the "that permanent's mana
// value" referent variant with a gain recipient: "Destroy target artifact. You
// gain life equal to that permanent's mana value." (Divine Offering).
func TestLowerDivineOfferingManaValueLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Divine Offering",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target artifact. You gain life equal to that permanent's mana value.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	dyn := lifeRiderDynamic(t, gain.Amount)
	if dyn.Kind != game.DynamicAmountObjectManaValue || dyn.Object != game.TargetPermanentReference(0) {
		t.Fatalf("dynamic = %+v, want ObjectManaValue of target 0", dyn)
	}
}

func TestLowerReanimateManaValueLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reanimate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to that card's mana value.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("first primitive = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	card, ok := put.Source.CardRef()
	if !ok || card.Kind != game.CardReferenceTarget || card.TargetIndex != 0 {
		t.Fatalf("put source = %+v, want target card 0", put.Source)
	}
	if !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("put recipient = %+v, want controller", put.Recipient)
	}
	if put.PublishLinked == "" || mode.Sequence[0].PublishResult == "" {
		t.Fatalf("put instruction = %+v, want linked permanent and result publication", mode.Sequence[0])
	}
	lose, ok := mode.Sequence[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.LoseLife", mode.Sequence[1].Primitive)
	}
	dyn := lifeRiderDynamic(t, lose.Amount)
	if dyn.Kind != game.DynamicAmountObjectManaValue ||
		dyn.Object != game.LinkedObjectReference(string(put.PublishLinked)) {
		t.Fatalf("dynamic = %+v, want mana value of linked permanent", dyn)
	}
	gate := mode.Sequence[1].ResultGate
	if !gate.Exists ||
		gate.Val.Key != mode.Sequence[0].PublishResult ||
		gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("life rider gate = %+v, want successful-move gate", gate)
	}
}

func TestLowerGraveyardReturnManaValueGainLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Restorative Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card from a graveyard to the battlefield under your control. You gain life equal to that card's mana value.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %+v, want two instructions", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok || put.PublishLinked == "" {
		t.Fatalf("first primitive = %+v, want linked PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	dyn := lifeRiderDynamic(t, gain.Amount)
	if dyn.Kind != game.DynamicAmountObjectManaValue ||
		dyn.Object != game.LinkedObjectReference(string(put.PublishLinked)) {
		t.Fatalf("dynamic = %+v, want mana value of linked permanent", dyn)
	}
}

// TestLifeRiderFailsClosed verifies the rider stays fail-closed for amount
// characteristics and recipients it does not model: a mana-value amount paired
// with exile (which cannot prove the destroyed-permanent referent), unsupported
// graveyard-return shapes, and a lone life clause with no prior subject.
func TestLifeRiderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// Mana value is only modeled when a prior clause destroys the target
		// permanent; exile cannot prove the battlefield referent the same way.
		"Exile target creature. Its controller gains life equal to its mana value.",
		// "Its" binds directly to a target, but the target is a card rather than
		// a battlefield object; only the exact prior-result wording is supported.
		"Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to its mana value.",
		"Put target artifact card from a graveyard onto the battlefield under your control. You lose life equal to that card's mana value.",
		"Put up to two target creature cards from graveyards onto the battlefield under your control. You lose life equal to that card's mana value.",
		"Put target creature card from a graveyard into its owner's hand. You lose life equal to that card's mana value.",
		"Put target creature card from a graveyard onto the battlefield under its owner's control. You lose life equal to that card's mana value.",
		"You may put target creature card from a graveyard onto the battlefield under your control. You lose life equal to that card's mana value.",
		"Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to that card's power.",
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
