package parser

import "testing"

func TestRecognizeClassLevelGain(t *testing.T) {
	const oracle = "(Gain the next level as a sorcery to add its ability.)\n" +
		"You have no maximum hand size.\n" +
		"{2}{U}: Level 2\n" +
		"Whenever you draw a card, put a +1/+1 counter on target creature you control.\n" +
		"{4}{U}: Level 3\n" +
		"Creatures you control have haste."
	document, diagnostics := Parse(oracle, Context{Class: true, CardName: "Wizard Class"})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %v", diagnostics)
	}
	gains := map[int]int{}
	reminders := 0
	for i := range document.Abilities {
		ability := document.Abilities[i]
		if ability.ClassReminder {
			reminders++
		}
		if ability.ClassLevelGain != 0 {
			gains[i] = ability.ClassLevelGain
		}
	}
	if reminders != 1 {
		t.Fatalf("class reminders = %d, want 1", reminders)
	}
	if len(gains) != 2 {
		t.Fatalf("class level-up abilities = %d, want 2 (%v)", len(gains), gains)
	}
	for index, level := range gains {
		if document.Abilities[index].Kind != AbilityActivated {
			t.Fatalf("ability %d with level gain has kind %v, want AbilityActivated", index, document.Abilities[index].Kind)
		}
		if level != 2 && level != 3 {
			t.Fatalf("unexpected level-up target %d", level)
		}
	}
}

// TestRecognizeClassLevelGainRequiresClassContext verifies the level-up wording
// is recognized only for Class cards, so an ordinary "{cost}: Level N" body on a
// non-Class card stays unrecognized.
func TestRecognizeClassLevelGainRequiresClassContext(t *testing.T) {
	const oracle = "{2}{U}: Level 2"
	document, _ := Parse(oracle, Context{CardName: "Not A Class"})
	for i := range document.Abilities {
		if document.Abilities[i].ClassLevelGain != 0 {
			t.Fatalf("ability %d recognized a level gain without Class context", i)
		}
	}
}
