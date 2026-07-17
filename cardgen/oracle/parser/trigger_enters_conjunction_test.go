package parser

import "testing"

func TestExpandEntersAndSecondTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "enters and upkeep phase trigger",
			source: "When this creature enters and at the beginning of your upkeep, surveil 1.",
			want: []string{
				"When this creature enters, surveil 1.",
				"At the beginning of your upkeep, surveil 1.",
			},
		},
		{
			name:   "enters and when-you-sacrifice trigger shares effect body",
			source: "When this artifact enters and when you sacrifice it, create a 1/1 white Rabbit creature token and scry 1.",
			want: []string{
				"When this artifact enters, create a 1/1 white Rabbit creature token and scry 1.",
				"When you sacrifice it, create a 1/1 white Rabbit creature token and scry 1.",
			},
		},
		{
			name:   "enters and whenever-you-gain-life trigger",
			source: "When MACH-1 enters and whenever you gain life, surveil 1.",
			want: []string{
				"When MACH-1 enters, surveil 1.",
				"Whenever you gain life, surveil 1.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Aberrant Tinkering — When this creature enters and at the beginning of your upkeep, draw a card.",
			want: []string{
				"Aberrant Tinkering — When this creature enters, draw a card.",
				"Aberrant Tinkering — At the beginning of your upkeep, draw a card.",
			},
		},
		{
			name:   "exact comma-bearing card name and battlefield wording",
			source: "When Minsc & Boo, Timeless Heroes enters the battlefield and at the beginning of your upkeep, draw a card.",
			want: []string{
				"When Minsc & Boo, Timeless Heroes enters the battlefield, draw a card.",
				"At the beginning of your upkeep, draw a card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandEntersAndSecondTrigger(test.source, "Minsc & Boo, Timeless Heroes")
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandEntersAndSecondTrigger(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandEntersAndSecondTrigger(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandEntersAndSecondTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature enters, draw a card.",
		"At the beginning of your upkeep, draw a card.",
		// Second clause is an event subject, not a recognized trigger lead-in.
		"When this creature enters and another creature you control dies, draw a card.",
		// No effect body after the second condition.
		"When this creature enters and at the beginning of your upkeep",
		// Subject before "enters and" carries a comma.
		"When Old Rutstein, the Grim enters and at the beginning of your upkeep, mill a card.",
	}
	for _, source := range unchanged {
		if got := expandEntersAndSecondTrigger(source); got != source {
			t.Fatalf("expandEntersAndSecondTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
