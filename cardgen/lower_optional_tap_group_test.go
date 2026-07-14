package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// myrBattlesphereText is Myr Battlesphere's full Oracle text: an enter trigger
// that creates four Myr tokens and an attack trigger that optionally taps any
// number of untapped Myr the controller owns to pump the source and burn the
// attacked defender for the number tapped.
const myrBattlesphereText = "When this creature enters, create four 1/1 colorless Myr artifact creature tokens.\n" +
	"Whenever this creature attacks, you may tap X untapped Myr you control. If you do, this creature gets +X/+0 until end of turn and deals X damage to the player or planeswalker it's attacking."

// attackTrigger returns the single EventAttackerDeclared triggered ability in
// the lowered face, failing the test when it is absent.
func attackTrigger(t *testing.T, face loweredFaceAbilities) game.TriggeredAbility {
	t.Helper()
	for _, ability := range face.TriggeredAbilities {
		if ability.Trigger.Pattern.Event == game.EventAttackerDeclared {
			return ability
		}
	}
	t.Fatalf("no EventAttackerDeclared trigger among %d triggered abilities", len(face.TriggeredAbilities))
	return game.TriggeredAbility{}
}

// TestLowerOptionalTapGroupScaledConsequence proves Myr Battlesphere's attack
// trigger lowers to the reusable optional-tap-group sequence: a TapChosenGroup
// that publishes the number of untapped Myr the controller tapped, a source
// +X/+0 pump gated on the tap, and X damage to the attacked defender gated on
// the tap, where X is the published tap count reused by both consequences.
func TestLowerOptionalTapGroupScaledConsequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Battlesphere",
		Layout:     "normal",
		ManaCost:   "{7}",
		TypeLine:   "Artifact Creature — Myr Construct",
		OracleText: myrBattlesphereText,
		Power:      new("4"),
		Toughness:  new("7"),
	})

	ability := attackTrigger(t, face)
	if len(ability.Content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(ability.Content.Modes))
	}
	sequence := ability.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence len = %d, want 3", len(sequence))
	}

	tap, ok := sequence[0].Primitive.(game.TapChosenGroup)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.TapChosenGroup", sequence[0].Primitive)
	}
	if tap.PublishCount != optionalTapGroupCountKey {
		t.Fatalf("tap publish count = %q, want %q", tap.PublishCount, optionalTapGroupCountKey)
	}
	if sequence[0].PublishResult != optionalTapGroupCountKey {
		t.Fatalf("sequence[0] publish result = %q, want %q", sequence[0].PublishResult, optionalTapGroupCountKey)
	}
	wantGroup := game.PlayerControlledGroup(game.ControllerReference(), game.Selection{
		SubtypesAny: []types.Sub{types.Sub("Myr")},
		Controller:  game.ControllerYou,
		Tapped:      game.TriFalse,
	})
	if !reflect.DeepEqual(tap.ChooseFrom, wantGroup) {
		t.Fatalf("tap group = %#v, want %#v", tap.ChooseFrom, wantGroup)
	}

	wantAmount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: optionalTapGroupCountKey,
	})
	wantGate := game.InstructionResultGate{Key: optionalTapGroupCountKey, Succeeded: game.TriTrue}

	pump, ok := sequence[1].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.ModifyPT", sequence[1].Primitive)
	}
	if pump.Object != game.SourcePermanentReference() {
		t.Fatalf("pump object = %#v, want source permanent", pump.Object)
	}
	if !reflect.DeepEqual(pump.PowerDelta, wantAmount) {
		t.Fatalf("pump power delta = %#v, want chosen number", pump.PowerDelta)
	}
	if pump.ToughnessDelta != game.Fixed(0) {
		t.Fatalf("pump toughness delta = %#v, want fixed 0", pump.ToughnessDelta)
	}
	if pump.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("pump duration = %#v, want until end of turn", pump.Duration)
	}
	if !sequence[1].ResultGate.Exists || sequence[1].ResultGate.Val != wantGate {
		t.Fatalf("pump result gate = %#v, want %#v", sequence[1].ResultGate, wantGate)
	}

	damage, ok := sequence[2].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("sequence[2] = %T, want game.Damage", sequence[2].Primitive)
	}
	if !reflect.DeepEqual(damage.Amount, wantAmount) {
		t.Fatalf("damage amount = %#v, want chosen number", damage.Amount)
	}
	if damage.Recipient != game.AttackedDefenderDamageRecipient() {
		t.Fatalf("damage recipient = %#v, want defending player", damage.Recipient)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage source = %#v, want source permanent", damage.DamageSource)
	}
	if !sequence[2].ResultGate.Exists || sequence[2].ResultGate.Val != wantGate {
		t.Fatalf("damage result gate = %#v, want %#v", sequence[2].ResultGate, wantGate)
	}
}

// TestLowerOptionalTapGroupUnsupportedConsequenceFailsClosed proves the
// optional-tap-group recognizer fails closed when the "If you do" consequence is
// not one of the supported scaled effects (here, drawing cards), leaving the
// ability unsupported rather than lowering a partial or wrong sequence.
func TestLowerOptionalTapGroupUnsupportedConsequenceFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Draw Battlesphere",
		Layout:     "normal",
		ManaCost:   "{7}",
		TypeLine:   "Artifact Creature — Myr Construct",
		OracleText: "Whenever this creature attacks, you may tap X untapped Myr you control. If you do, draw X cards.",
		Power:      new("4"),
		Toughness:  new("7"),
	})
}
