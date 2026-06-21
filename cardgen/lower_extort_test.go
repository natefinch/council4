package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerExtortKeyword verifies that the printed Extort keyword expands to the
// optional "you may pay {W/B}" spell-cast trigger and lowers to a payment-gated
// drain: each opponent loses 1 life and the controller gains that much, both
// gated on the optional payment.
func TestLowerExtortKeyword(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "3"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Syndic of Tithes",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{2}{W}",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each opponent loses 1 life and you gain that much life.)",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventSpellCast ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger pattern = %#v, want a you-cast-a-spell trigger", trigger.Trigger.Pattern)
	}
	seq := trigger.Content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %#v, want pay + lose + gain", seq)
	}
	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.Pay", seq[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists || seq[0].PublishResult != controllerPaidResultKey {
		t.Fatalf("payment instruction = %#v, want a mana payment publishing %q", seq[0], controllerPaidResultKey)
	}
	lose, ok := seq[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("instruction[1] = %T, want game.LoseLife", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists || seq[1].ResultGate.Val.Key != controllerPaidResultKey {
		t.Fatalf("lose instruction gate = %#v, want gate on %q", seq[1].ResultGate, controllerPaidResultKey)
	}
	_ = lose
	gain, ok := seq[2].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("instruction[2] = %T, want game.GainLife", seq[2].Primitive)
	}
	if !seq[2].ResultGate.Exists || seq[2].ResultGate.Val.Key != controllerPaidResultKey {
		t.Fatalf("gain instruction gate = %#v, want gate on %q", seq[2].ResultGate, controllerPaidResultKey)
	}
	_ = gain
}

// TestLowerOptionalPaidBenefitNonControllerBody verifies the generic optional
// "you may pay {mana}. If you do, <body>" rider lowers a non-controller drain
// body in any spell-cast trigger, not only the printed Extort keyword.
func TestLowerOptionalPaidBenefitNonControllerBody(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Toll Collector",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		ManaCost:   "{1}{B}",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Whenever you cast a spell, you may pay {2}. If you do, each opponent loses 1 life and you gain that much life.",
	})
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence = %#v, want pay + lose + gain", seq)
	}
	if _, ok := seq[0].Primitive.(game.Pay); !ok {
		t.Fatalf("instruction[0] = %T, want game.Pay", seq[0].Primitive)
	}
}
