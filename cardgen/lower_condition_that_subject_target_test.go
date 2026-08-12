package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerDeathsCaressPastTenseNoTriggerBindsTarget guards the compiler's
// no-trigger fallback in bindConditionReferences with a real card: "Destroy
// target creature. If that creature was a Human, you gain life equal to its
// toughness." (Death's Caress). recognizeThatSubjectMatchCondition parses
// "that creature was a Human" as bound to the triggering event's permanent
// (its only binding today, needed for already-shipped cards like Endless
// Evil and Venom, Eddie Brock -- see TestLowerOtherwiseBranchKeyedOnEventPower
// and this file's own past-tense-with-trigger test below), but Death's
// Caress's ability has no trigger at all, so that binding could never have
// been correct; the compiler rebinds it to the ability's sole target
// instead, read through last-known information now that the destroy has
// resolved (CR 608.2b).
func TestLowerDeathsCaressPastTenseNoTriggerBindsTarget(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Death's Caress Test",
		"Destroy target creature. If that creature was a Human, you gain life equal to its toughness.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions (destroy, gated life gain)", sequence)
	}
	destroy := sequence[0]
	if _, ok := destroy.Primitive.(game.Destroy); !ok {
		t.Fatalf("instruction[0] = %T, want game.Destroy", destroy.Primitive)
	}
	gainLife := sequence[1]
	if _, ok := gainLife.Primitive.(game.GainLife); !ok {
		t.Fatalf("instruction[1] = %T, want game.GainLife", gainLife.Primitive)
	}
	gate := effectConditionMatch(t, gainLife)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target permanent 0", gate.Object)
	}
	if !slices.Contains(gate.ObjectMatches.Val.SubtypesAny, types.Sub("Human")) {
		t.Fatalf("gate selection = %#v, want subtype Human", gate.ObjectMatches.Val)
	}
}

// TestLowerPastTenseThatSubjectWithTriggerStaysEventPermanent guards that the
// no-trigger fallback above does NOT engage when a trigger genuinely does
// provide an event permanent, following the real-card shape
// TestLowerOtherwiseBranchKeyedOnEventPower already covers for a different
// predicate (Overgrowth Elemental: "Whenever another creature you control
// dies, you gain 1 life. If that creature was an Elemental, ..."),
// simplified to a gated draw so the test isolates the condition-binding gate
// from an unrelated counter-placement blocker. The dying creature is the
// EventPermanent, not a target -- this ability has no target at all -- so
// the fallback's own "trigger == nil" guard must correctly leave it alone.
func TestLowerPastTenseThatSubjectWithTriggerStaysEventPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Overgrowth Elemental Test",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		ManaCost:   "{3}{G}",
		OracleText: "Whenever another creature you control dies, you gain 1 life. If that creature was an Elemental, draw a card.",
		Power:      new("4"),
		Toughness:  new("4"),
	})
	var mode game.Mode
	found := false
	for _, ability := range face.TriggeredAbilities {
		for _, m := range ability.Content.Modes {
			if len(m.Sequence) == 2 {
				mode = m
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("triggered abilities = %#v, want a two-instruction mode", face.TriggeredAbilities)
	}
	gatedDraw := mode.Sequence[1]
	if _, ok := gatedDraw.Primitive.(game.Draw); !ok {
		t.Fatalf("sequence[1] = %T, want game.Draw", gatedDraw.Primitive)
	}
	gate := effectConditionMatch(t, gatedDraw)
	if gate.Object.Val.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("gate object = %#v, want event permanent", gate.Object)
	}
}

// TestLowerPresentTenseThatSubjectBindsTarget guards the present-tense
// recognizer (recognizeThatSubjectTargetMatchCondition) with a card
// following the exact grammar of real cards this family covers (Splash
// Portal: "... If that creature is a Bird, Frog, Otter, or Rat, draw a
// card."; Yip Yip!: "... If that creature is an Ally, ..."), simplified to
// isolate the condition gate from those two cards' own separate, unrelated
// blockers (an "under its owner's control" blink-return reference-binding
// gap for Splash Portal's own exact wording -- see
// TestLowerBlinkThatSubjectStillFailsClosed below -- and an "until end of
// turn" temporary-keyword-grant gap for Yip Yip's own exact wording).
func TestLowerPresentTenseThatSubjectBindsTarget(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Present Tense That Subject Test",
		"Target creature gets +2/+2 until end of turn. If that creature is an Ally, you draw a card.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions (pump, gated draw)", sequence)
	}
	pump := sequence[0]
	if _, ok := pump.Primitive.(game.ModifyPT); !ok {
		t.Fatalf("instruction[0] = %T, want game.ModifyPT", pump.Primitive)
	}
	draw := sequence[1]
	if _, ok := draw.Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[1] = %T, want game.Draw", draw.Primitive)
	}
	gate := effectConditionMatch(t, draw)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target permanent 0", gate.Object)
	}
	if !slices.Contains(gate.ObjectMatches.Val.SubtypesAny, types.Sub("Ally")) {
		t.Fatalf("gate selection = %#v, want subtype Ally", gate.ObjectMatches.Val)
	}
}

// TestLowerBareItsRequiredTypeNoTriggerBindsTarget guards that the compiler's
// no-trigger fallback also generalizes to the pre-existing bare-pronoun "it's
// a <selection>" recognizer (recognizeEventSubjectMatchCondition), not just
// the new "that <noun>" recognizers above, using the real card Blacksmith's
// Skill: "Target permanent gains hexproof and indestructible until end of
// turn. If it's an artifact creature, it gets +2/+2 until end of turn."
// "Artifact creature" has two required card types, which
// recognizeTargetObjectMatchCondition (the OTHER "it's a <selection>"
// recognizer, restricted to subtype/supertype-only selections) explicitly
// declines, so this clause can only have reached
// recognizeEventSubjectMatchCondition's EventPermanent-bound fallback path --
// confirming the compiler fallback's generality was not accidental to this
// family's own new recognizer.
func TestLowerBareItsRequiredTypeNoTriggerBindsTarget(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Blacksmith's Skill",
		"Target permanent gains hexproof and indestructible until end of turn. If it's an artifact creature, it gets +2/+2 until end of turn.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions (keyword grant, gated pump)", sequence)
	}
	pump := sequence[1]
	if _, ok := pump.Primitive.(game.ModifyPT); !ok {
		t.Fatalf("instruction[1] = %T, want game.ModifyPT", pump.Primitive)
	}
	gate := effectConditionMatch(t, pump)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target permanent 0", gate.Object)
	}
	if !slices.Contains(gate.ObjectMatches.Val.RequiredTypes, types.Artifact) ||
		!slices.Contains(gate.ObjectMatches.Val.RequiredTypes, types.Creature) {
		t.Fatalf("gate selection = %#v, want required types [Artifact Creature]", gate.ObjectMatches.Val)
	}
}

// TestLowerBlinkThatSubjectStillFailsClosed guards against a genuine
// ambiguity this family's development uncovered: Splash Portal's exact
// wording ("Exile target creature you control, then return it to the
// battlefield under its owner's control. If that creature is a Bird, Frog,
// Otter, or Rat, draw a card.") still fails closed, because "that creature"
// here does not name the ORIGINAL target -- CR 400.7 makes the blinked
// permanent a new object once it returns, so "that creature" names the
// just-returned object (a prior-instruction-result / linked-object
// antecedent the general reference machinery already correctly resolves),
// not literally game.ObjectReferenceTargetPermanent. The parser's present-
// tense recognizer still guesses Target (a reasonable default for the
// non-blink shapes above), but the compiler's existing conflict-detection
// safety net (conditionObjectBinding, unchanged by this family's work) finds
// the reference machinery's own resolution disagrees and fails closed rather
// than picking either one -- exactly the same fail-safe that made the
// no-trigger Target fallback above safe to add in the first place. This
// blink-family variant is a real, documented follow-up (needs
// ConditionObjectBindingPriorInstructionResult support for this predicate),
// not a bug in this slice.
func TestLowerBlinkThatSubjectStillFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Splash Portal Probe",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Rat, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 && source != "" {
		t.Fatal("blink that-subject gate unexpectedly compiled without any diagnostic")
	}
}
