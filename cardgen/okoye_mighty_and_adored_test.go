package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceOkoyeMightyAndAdored covers the
// beginning-of-combat trigger that puts a +1/+1 counter on target creature and
// then sets up a delayed "attacks the monarch this turn" trigger on that same
// creature. The counter placement publishes the target under a linked key
// (AddCounter.PublishLinked) and the delayed trigger binds that key
// (CapturedAttackerObject) so it fires only when that specific creature is
// declared attacking the monarch (Player: TriggerPlayerMonarch,
// AttackRecipient: AttackRecipientPlayer, AttackerCaptured: true), granting it
// double strike and trample until end of turn.
func TestGenerateExecutableCardSourceOkoyeMightyAndAdored(t *testing.T) {
	t.Parallel()
	p, tt := "3", "3"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Okoye, Mighty and Adored",
		Layout:    "normal",
		ManaCost:  "{2}{G}{W}",
		TypeLine:  "Legendary Creature — Human Warrior Hero",
		Power:     &p,
		Toughness: &tt,
		OracleText: "When Okoye enters, you become the monarch.\n" +
			"At the beginning of combat on your turn, put a +1/+1 counter on target creature. Whenever that creature attacks the monarch this turn, it gains double strike and trample until end of turn.",
	}, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// The counter placement publishes the pumped creature under a linked key.
		"Primitive: game.AddCounter{",
		"CounterKind:   counter.PlusOnePlusOne",
		`PublishLinked: game.LinkedKey("delayed-target-1")`,
		// The delayed trigger binds that key and fires on an attacks-the-monarch
		// declaration by that captured creature.
		"Primitive: game.CreateDelayedTrigger{",
		"Event:            game.EventAttackerDeclared",
		"Player:           game.TriggerPlayerMonarch",
		"AttackerCaptured: true",
		"AttackRecipient:  game.AttackRecipientPlayer",
		"Window:                 game.DelayedWindowThisTurn",
		`CapturedAttackerObject: opt.Val(game.LinkedObjectReference("delayed-target-1"))`,
		// The granted keywords apply to the attacking creature until end of turn.
		"Object: opt.Val(game.EventPermanentReference())",
		"game.DoubleStrike",
		"game.Trample",
		"Duration: game.DurationUntilEndOfTurn",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The delayed trigger must not be a static source-filtered attacker trigger.
	if strings.Contains(source, "Source:           game.TriggerSourceSelf") {
		t.Fatalf("delayed trigger kept a self source filter instead of binding the captured attacker:\n%s", source)
	}
}
