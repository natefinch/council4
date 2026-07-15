package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerFiremaneCommando covers the reusable attack-batch trigger semantics.
// The card grants Flying plus two attacker-declared triggers: the controller's
// own "attack with two or more creatures" draw, and the opponent-scoped
// "Whenever another player attacks with two or more creatures, they draw a card
// if none of those creatures attacked you." The second trigger lowers to an
// event-player draw gated by the no-attacker-attacked-controller condition, so
// the attacking player draws only when their batch declared no direct attack on
// this card's controller.
func TestLowerFiremaneCommando(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Firemane Commando",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{3}{W}",
		OracleText: "Flying\nWhenever you attack with two or more creatures, draw a card.\nWhenever another player attacks with two or more creatures, they draw a card if none of those creatures attacked you.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}

	own := face.TriggeredAbilities[0].Trigger.Pattern
	if own.Event != game.EventAttackerDeclared ||
		own.Controller != game.TriggerControllerYou ||
		own.AttackerCountAtLeast != 2 ||
		!own.OneOrMore {
		t.Fatalf("own attack trigger = %#v", own)
	}

	other := face.TriggeredAbilities[1]
	pattern := other.Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Controller != game.TriggerControllerOpponent ||
		pattern.AttackerCountAtLeast != 2 ||
		!pattern.OneOrMore {
		t.Fatalf("other-player attack trigger = %#v", pattern)
	}
	if other.Trigger.InterveningCondition.Exists {
		t.Fatalf("gate must be a per-effect condition, not intervening: %#v", other.Trigger.InterveningCondition)
	}

	if len(other.Content.Modes) != 1 || len(other.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("other content = %#v", other.Content)
	}
	instruction := other.Content.Modes[0].Sequence[0]
	draw, ok := instruction.Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %#v, want Draw", instruction.Primitive)
	}
	if draw.Player.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("draw player = %#v, want event player", draw.Player)
	}
	if !instruction.Condition.Exists {
		t.Fatalf("draw must be gated by a condition: %#v", instruction)
	}
	gate := instruction.Condition.Val.Condition
	if !gate.Exists || len(gate.Val.Aggregates) != 1 ||
		gate.Val.Aggregates[0].Aggregate != game.AggregateAttackersInBatchAttackedController {
		t.Fatalf("gate condition = %#v, want attackers-in-batch-attacked-controller aggregate", gate)
	}
}
