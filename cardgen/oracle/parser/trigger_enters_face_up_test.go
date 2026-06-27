package parser

import "testing"

func TestExpandEntersOrTurnedFaceUpTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "self trigger splits into enters and turned-face-up triggers",
			source: "When this creature enters or is turned face up, create three 1/1 red Goblin creature tokens.",
			want: []string{
				"When this creature enters, create three 1/1 red Goblin creature tokens.",
				"Whenever this creature is turned face up, create three 1/1 red Goblin creature tokens.",
			},
		},
		{
			name:   "whenever introduction preserved on enters half",
			source: "Whenever this creature enters or is turned face up, draw a card.",
			want: []string{
				"Whenever this creature enters, draw a card.",
				"Whenever this creature is turned face up, draw a card.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Disguise — When this creature enters or is turned face up, draw a card.",
			want: []string{
				"Disguise — When this creature enters, draw a card.",
				"Disguise — Whenever this creature is turned face up, draw a card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandEntersOrTurnedFaceUpTrigger(test.source)
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandEntersOrTurnedFaceUpTrigger(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandEntersOrTurnedFaceUpTrigger(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandEntersOrTurnedFaceUpTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature enters, draw a card.",
		"Whenever this creature is turned face up, draw a card.",
		"When this creature enters or is turned face up",
		"When this creature enters or is turned face up,",
		"Megamorph {3}{G} (You may cast this card face down as a 2/2 creature for {3}.)",
	}
	for _, source := range unchanged {
		if got := expandEntersOrTurnedFaceUpTrigger(source); got != source {
			t.Fatalf("expandEntersOrTurnedFaceUpTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
