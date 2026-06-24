package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDiscardTriggerGroupDamageThatMuch asserts "deals that much damage to
// each opponent." inside a discard trigger resolves the triggering discarded
// card count as group-wide damage (Magmakin Artillerist's first ability).
func TestLowerDiscardTriggerGroupDamageThatMuch(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Discard Gunner",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Elemental",
		OracleText: "Whenever you discard one or more cards, this creature deals that much damage to each opponent.",
		Colors:     []string{"R"},
		Power:      new("1"),
		Toughness:  new("4"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventCardDiscarded || !ta.Trigger.Pattern.OneOrMore {
		t.Fatalf("trigger pattern = %#v", ta.Trigger.Pattern)
	}
	damage, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Damage", ta.Content.Modes[0].Sequence[0].Primitive)
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventCardCount {
		t.Fatalf("damage amount = %#v, want dynamic event card count", damage.Amount)
	}
	if _, ok := damage.Recipient.PlayerGroupReference(); !ok {
		t.Fatalf("damage recipient = %#v, want each-opponent player group", damage.Recipient)
	}
}

// TestLowerDiscardTriggerDrawThatMany asserts "draw that many cards." inside a
// discard trigger resolves the triggering discarded card count as the draw
// amount, independent of the historical anaphor kind the parser pins (Rielle,
// the Everwise's draw rider).
func TestLowerDiscardTriggerDrawThatMany(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Discard Sage",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you discard one or more cards, draw that many cards.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventCardDiscarded || !ta.Trigger.Pattern.OneOrMore {
		t.Fatalf("trigger pattern = %#v", ta.Trigger.Pattern)
	}
	draw, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Draw", ta.Content.Modes[0].Sequence[0].Primitive)
	}
	dynamic := draw.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventCardCount {
		t.Fatalf("draw amount = %#v, want dynamic event card count", draw.Amount)
	}
}

// TestLowerDamageTriggerGroupDamageThatMuchStillCounterEvent guards against a
// regression: in a counter-placement trigger the same "that much" anaphor must
// still resolve to the counter count, not the card count.
func TestLowerDamageTriggerDealtThatMuch(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spite Mirror",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Elemental",
		OracleText: "Whenever Spite Mirror is dealt damage, it deals that much damage to any target.",
		Colors:     []string{"R"},
		Power:      new("0"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventDamageDealt {
		t.Fatalf("trigger pattern event = %v, want EventDamageDealt", ta.Trigger.Pattern.Event)
	}
	damage, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Damage", ta.Content.Modes[0].Sequence[0].Primitive)
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
		t.Fatalf("damage amount = %#v, want dynamic event damage", damage.Amount)
	}
}
