package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// insteadEscalationModes pulls the two gated instructions out of a single-mode
// spell body that lowered a "[base]. If <condition>, [escalated] instead."
// escalation, failing the test on any other shape.
func insteadEscalationModes(t *testing.T, content game.AbilityContent) (base, escalated game.Instruction) {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %#v", content.Modes)
	}
	seq := content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two gated instructions", seq)
	}
	return seq[0], seq[1]
}

// TestLowerInsteadEscalationLifeGain proves the trailing-"instead" escalation
// exactness generalizes beyond monarch and token creation: "You gain 4 life. If
// a creature died this turn, you gain 8 life instead." (Life Goes On) lowers to
// two GainLife instructions gated on the negated and plain event-history
// condition, so exactly one resolves.
func TestLowerInsteadEscalationLifeGain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Life Goes On",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You gain 4 life. If a creature died this turn, you gain 8 life instead.",
	})
	base, escalated := insteadEscalationModes(t, face.SpellAbility.Val)
	baseGain, ok := base.Primitive.(game.GainLife)
	if !ok || baseGain.Amount != game.Fixed(4) {
		t.Fatalf("base = %#v, want GainLife 4", base.Primitive)
	}
	escGain, ok := escalated.Primitive.(game.GainLife)
	if !ok || escGain.Amount != game.Fixed(8) {
		t.Fatalf("escalation = %#v, want GainLife 8", escalated.Primitive)
	}
	if !base.Condition.Exists || !base.Condition.Val.Condition.Exists ||
		!base.Condition.Val.Condition.Val.Negate {
		t.Fatalf("base gate = %#v, want negated condition", base.Condition)
	}
	if !escalated.Condition.Exists || !escalated.Condition.Val.Condition.Exists ||
		escalated.Condition.Val.Condition.Val.Negate {
		t.Fatalf("escalation gate = %#v, want plain condition", escalated.Condition)
	}
}

// TestLowerInsteadEscalationDraw proves the draw-spell "instead" escalation
// ("Draw two cards. If this spell was kicked, draw three cards instead.", Field
// Research) lowers both branches, confirming the exactness fix composes across
// effect types.
func TestLowerInsteadEscalationDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Field Research",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Kicker {2}{U} (You may pay an additional {2}{U} as you cast this spell.)\n" +
			"Draw two cards. If this spell was kicked, draw three cards instead.",
	})
	base, escalated := insteadEscalationModes(t, face.SpellAbility.Val)
	if _, ok := base.Primitive.(game.Draw); !ok {
		t.Fatalf("base = %#v, want Draw", base.Primitive)
	}
	if _, ok := escalated.Primitive.(game.Draw); !ok {
		t.Fatalf("escalation = %#v, want Draw", escalated.Primitive)
	}
	if !base.Condition.Exists || !base.Condition.Val.Condition.Exists ||
		!base.Condition.Val.Condition.Val.Negate {
		t.Fatalf("base gate = %#v, want negated condition", base.Condition)
	}
	if !escalated.Condition.Exists || !escalated.Condition.Val.Condition.Exists ||
		escalated.Condition.Val.Condition.Val.Negate {
		t.Fatalf("escalation gate = %#v, want plain condition", escalated.Condition)
	}
}
