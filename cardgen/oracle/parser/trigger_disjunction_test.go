package parser

import "testing"

func TestExpandDisjunctiveTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "three event union shares effect",
			source: "Whenever another creature dies, or a creature card is put into a graveyard from anywhere other than the battlefield, or a creature card leaves your graveyard, Syr Konrad deals 1 damage to each opponent.",
			want: []string{
				"Whenever another creature dies, Syr Konrad deals 1 damage to each opponent.",
				"Whenever a creature card is put into a graveyard from anywhere other than the battlefield, Syr Konrad deals 1 damage to each opponent.",
				"Whenever a creature card leaves your graveyard, Syr Konrad deals 1 damage to each opponent.",
			},
		},
		{
			name:   "two event union shares effect",
			source: "Whenever another creature dies, or a creature card leaves your graveyard, this creature deals 1 damage to each opponent.",
			want: []string{
				"Whenever another creature dies, this creature deals 1 damage to each opponent.",
				"Whenever a creature card leaves your graveyard, this creature deals 1 damage to each opponent.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Eerie — Whenever an enchantment you control enters, or a creature card leaves your graveyard, draw a card.",
			want: []string{
				"Eerie — Whenever an enchantment you control enters, draw a card.",
				"Eerie — Whenever a creature card leaves your graveyard, draw a card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandDisjunctiveTrigger(test.source)
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandDisjunctiveTrigger(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandDisjunctiveTrigger(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandDisjunctiveTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature enters, draw a card.",
		"Whenever you cast a spell, you may pay {W/B}.",
		"Choose one — Draw a card, or gain 2 life.",
		"Whenever this creature attacks, it gains first strike, or trample until end of turn.",
		"When this creature dies, or whatever",
	}
	for _, source := range unchanged {
		if got := expandDisjunctiveTrigger(source); got != source {
			t.Fatalf("expandDisjunctiveTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
