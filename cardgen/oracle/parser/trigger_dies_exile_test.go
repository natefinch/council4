package parser

import "testing"

func TestExpandDiesOrExileTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "reanimation aura splits into death and exile triggers",
			source: "When enchanted permanent dies or is put into exile, return that card to the battlefield under your control.",
			want: []string{
				"When enchanted permanent dies, return that card to the battlefield under your control.",
				"Whenever enchanted permanent is put into exile, return that card to the battlefield under your control.",
			},
		},
		{
			name:   "from the battlefield qualifier preserved on exile half",
			source: "Whenever enchanted creature dies or is put into exile from the battlefield, draw a card.",
			want: []string{
				"Whenever enchanted creature dies, draw a card.",
				"Whenever enchanted creature is put into exile from the battlefield, draw a card.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Eerie — When enchanted permanent dies or is put into exile, draw a card.",
			want: []string{
				"Eerie — When enchanted permanent dies, draw a card.",
				"Eerie — Whenever enchanted permanent is put into exile, draw a card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandDiesOrExileTrigger(test.source)
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandDiesOrExileTrigger(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandDiesOrExileTrigger(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandDiesOrExileTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature dies, draw a card.",
		"Whenever enchanted creature is put into exile, draw a card.",
		"When enchanted permanent dies or is put into exile",
		"Sacrifice this: When enchanted permanent dies or is put into exile, draw a card.",
		"When enchanted permanent dies or is put into exile,",
		"Enchanted creature gets +1/+1 and dies or is put into exile, draw a card.",
	}
	for _, source := range unchanged {
		if got := expandDiesOrExileTrigger(source); got != source {
			t.Fatalf("expandDiesOrExileTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
