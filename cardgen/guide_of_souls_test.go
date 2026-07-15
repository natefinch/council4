package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// guideOfSoulsCard is the full Guide of Souls oracle text. Its second triggered
// ability exercises the reusable typed support this test suite covers: an
// optional energy payment on attack, a reflexive "when you do" trigger that
// selects its target only after payment, a compound multi-kind counter placement
// (two +1/+1 plus one flying), and a permanent additive subtype change to Angel.
func guideOfSoulsCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Guide of Souls",
		Layout:    "normal",
		TypeLine:  "Creature — Human Cleric",
		ManaCost:  "{W}",
		Power:     new("1"),
		Toughness: new("2"),
		OracleText: "Whenever another creature you control enters, you gain 1 life and get {E} (an energy counter).\n" +
			"Whenever you attack, you may pay {E}{E}{E}. When you do, put two +1/+1 counters and a flying counter on target attacking creature. It becomes an Angel in addition to its other types.",
	}
}

// TestGenerateGuideOfSoulsEntersTrigger proves the first triggered ability lowers
// the ordered "you gain 1 life and get {E}" consequence into a GainLife followed
// by an energy AddPlayerCounter on the controller, gated on another creature you
// control entering.
func TestGenerateGuideOfSoulsEntersTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, guideOfSoulsCard())
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	enters := face.TriggeredAbilities[0]
	if enters.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("enters trigger event = %v, want EventPermanentEnteredBattlefield", enters.Trigger.Pattern.Event)
	}
	if !enters.Trigger.Pattern.ExcludeSelf {
		t.Error("enters trigger should exclude the source (\"another creature\")")
	}
	seq := enters.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("enters sequence = %#v, want two instructions", seq)
	}
	gain, ok := seq[0].Primitive.(game.GainLife)
	if !ok || gain.Amount != game.Fixed(1) {
		t.Fatalf("first instruction = %#v, want GainLife 1", seq[0].Primitive)
	}
	energy, ok := seq[1].Primitive.(game.AddPlayerCounter)
	if !ok || energy.CounterKind != counter.Energy || energy.Amount != game.Fixed(1) {
		t.Fatalf("second instruction = %#v, want AddPlayerCounter Energy 1", seq[1].Primitive)
	}
}

// TestGenerateGuideOfSoulsAttackReflexivePayment proves the second triggered
// ability lowers into an optional energy payment whose success gates a reflexive
// trigger: the target attacking creature is chosen only when the reflexive
// trigger goes on the stack (after payment), and declining or failing to pay
// creates no reflexive trigger because the Pay publishes a result the
// CreateReflexiveTrigger instruction gates on.
func TestGenerateGuideOfSoulsAttackReflexivePayment(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, guideOfSoulsCard())
	attack := face.TriggeredAbilities[1]
	if attack.Trigger.Pattern.Event != game.EventAttackerDeclared {
		t.Fatalf("attack trigger event = %v, want EventAttackerDeclared", attack.Trigger.Pattern.Event)
	}
	seq := attack.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("attack sequence = %#v, want two instructions (Pay + CreateReflexiveTrigger)", seq)
	}

	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first instruction = %#v, want game.Pay", seq[0].Primitive)
	}
	if seq[0].PublishResult == "" {
		t.Fatal("Pay instruction must publish a result key for the reflexive gate")
	}
	costs := pay.Payment.AdditionalCosts
	if len(costs) != 1 || costs[0].Kind != cost.AdditionalEnergy || costs[0].Amount != 3 {
		t.Fatalf("payment additional costs = %#v, want one AdditionalEnergy of 3", costs)
	}

	reflex, ok := seq[1].Primitive.(game.CreateReflexiveTrigger)
	if !ok {
		t.Fatalf("second instruction = %#v, want game.CreateReflexiveTrigger", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists {
		t.Fatal("CreateReflexiveTrigger must be gated on the payment result")
	}
	if gate := seq[1].ResultGate.Val; gate.Key != seq[0].PublishResult || gate.Succeeded != game.TriTrue {
		t.Fatalf("result gate = %#v, want key %q succeeded", gate, seq[0].PublishResult)
	}

	inner := reflex.Trigger.Content.Modes[0]
	if len(inner.Targets) != 1 || inner.Targets[0].MinTargets != 1 || inner.Targets[0].MaxTargets != 1 {
		t.Fatalf("reflexive targets = %#v, want one target attacking creature", inner.Targets)
	}
	assertGuideReflexiveBody(t, inner.Sequence)
}

// assertGuideReflexiveBody verifies the reflexive consequence body places two
// +1/+1 counters, then a flying counter, then permanently makes the target an
// Angel — all addressing the single reflexive target permanent.
func assertGuideReflexiveBody(t *testing.T, seq []game.Instruction) {
	t.Helper()
	if len(seq) != 3 {
		t.Fatalf("reflexive body = %#v, want three instructions", seq)
	}

	plus, ok := seq[0].Primitive.(game.AddCounter)
	if !ok || plus.CounterKind != counter.PlusOnePlusOne || plus.Amount != game.Fixed(2) ||
		plus.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first instruction = %#v, want AddCounter +1/+1 x2 on target", seq[0].Primitive)
	}

	flying, ok := seq[1].Primitive.(game.AddCounter)
	if !ok || flying.CounterKind != counter.Flying || flying.Amount != game.Fixed(1) ||
		flying.Object != game.TargetPermanentReference(0) {
		t.Fatalf("second instruction = %#v, want AddCounter flying x1 on target", seq[1].Primitive)
	}

	become, ok := seq[2].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("third instruction = %#v, want game.ApplyContinuous", seq[2].Primitive)
	}
	if !become.Object.Exists || become.Object.Val != game.TargetPermanentReference(0) {
		t.Fatalf("become object = %#v, want target permanent reference", become.Object)
	}
	if become.Duration != game.DurationPermanent {
		t.Fatalf("become duration = %v, want permanent", become.Duration)
	}
	if len(become.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", become.ContinuousEffects)
	}
	effect := become.ContinuousEffects[0]
	if effect.Layer != game.LayerType {
		t.Fatalf("continuous layer = %v, want LayerType", effect.Layer)
	}
	if len(effect.AddSubtypes) != 1 || effect.AddSubtypes[0] != types.Sub("Angel") {
		t.Fatalf("add subtypes = %#v, want [Angel]", effect.AddSubtypes)
	}
	if len(effect.AddTypes) != 0 {
		t.Fatalf("add types = %#v, want none (Angel is a subtype)", effect.AddTypes)
	}
}
