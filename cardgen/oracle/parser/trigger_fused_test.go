package parser

import "testing"

func TestExpandFusedTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		source     string
		wantFirst  string
		wantSecond string
	}{
		{
			name:       "enters and whenever opponent draws",
			source:     "When this creature enters and whenever an opponent draws a card, this creature deals 1 damage to any target.",
			wantFirst:  "When this creature enters, this creature deals 1 damage to any target.",
			wantSecond: "Whenever an opponent draws a card, this creature deals 1 damage to any target.",
		},
		{
			name:       "whenever cast spell with mana value",
			source:     "When this enchantment enters and whenever you cast a spell with mana value 5 or greater, draw a card.",
			wantFirst:  "When this enchantment enters, draw a card.",
			wantSecond: "Whenever you cast a spell with mana value 5 or greater, draw a card.",
		},
		{
			name:       "both introductions whenever",
			source:     "Whenever this creature enters and whenever you cast a party spell, choose a party creature card in your hand.",
			wantFirst:  "Whenever this creature enters, choose a party creature card in your hand.",
			wantSecond: "Whenever you cast a party spell, choose a party creature card in your hand.",
		},
		{
			name:       "ability word prefix carried to both",
			source:     "Eerie — Whenever an enchantment you control enters and whenever you fully unlock a Room, draw a card.",
			wantFirst:  "Eerie — Whenever an enchantment you control enters, draw a card.",
			wantSecond: "Eerie — Whenever you fully unlock a Room, draw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandFusedTrigger(test.source)
			lines := splitSourceLines(got)
			if len(lines) != 2 {
				t.Fatalf("expandFusedTrigger(%q) = %q, want two lines", test.source, got)
			}
			if lines[0] != test.wantFirst || lines[1] != test.wantSecond {
				t.Fatalf("expandFusedTrigger(%q) = %q, want [%q %q]", test.source, got, test.wantFirst, test.wantSecond)
			}
		})
	}
}

func TestExpandFusedTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature enters, draw a card.",
		"Whenever you cast a spell, you may pay {W/B}.",
		"At the beginning of your upkeep and whenever you draw, gain 1 life.",
		"This creature enters and whenever it attacks, deal damage.",
		"When this creature enters and whenever you draw a card",
	}
	for _, source := range unchanged {
		if got := expandFusedTrigger(source); got != source {
			t.Fatalf("expandFusedTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
