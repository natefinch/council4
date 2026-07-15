package parser

import "testing"

func TestExpandSharedSubjectTriggerUnion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "attacks blocks or becomes target shares subject and effect",
			source: "Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.",
			want: []string{
				"Whenever this creature attacks, it deals damage equal to its power to each opponent.",
				"Whenever this creature blocks, it deals damage equal to its power to each opponent.",
				"Whenever this creature becomes the target of a spell, it deals damage equal to its power to each opponent.",
			},
		},
		{
			name:   "spell or ability final condition kept whole",
			source: "Whenever this creature attacks, blocks, or becomes the target of a spell or ability, draw a card.",
			want: []string{
				"Whenever this creature attacks, draw a card.",
				"Whenever this creature blocks, draw a card.",
				"Whenever this creature becomes the target of a spell or ability, draw a card.",
			},
		},
		{
			name:   "shared subject may be a controlled creature",
			source: "Whenever a creature you control attacks, blocks, or becomes the target of a spell, you gain 1 life.",
			want: []string{
				"Whenever a creature you control attacks, you gain 1 life.",
				"Whenever a creature you control blocks, you gain 1 life.",
				"Whenever a creature you control becomes the target of a spell, you gain 1 life.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Toy — Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.",
			want: []string{
				"Toy — Whenever this creature attacks, it deals damage equal to its power to each opponent.",
				"Toy — Whenever this creature blocks, it deals damage equal to its power to each opponent.",
				"Toy — Whenever this creature becomes the target of a spell, it deals damage equal to its power to each opponent.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandSharedSubjectTriggerUnion(test.source)
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandSharedSubjectTriggerUnion(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandSharedSubjectTriggerUnion(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandSharedSubjectTriggerUnionLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		// Two-verb combat unions have no comma list; the union parsers handle them.
		"Whenever this creature attacks or blocks, it deals damage equal to its power to each opponent.",
		// The ", or " here joins effect clauses, not trigger conditions.
		"Whenever this creature attacks, it gains first strike, or trample until end of turn.",
		// A comma list whose members are not recognized bare combat/target verbs.
		"Whenever this creature attacks, deals combat damage, or dies, draw a card.",
		// No shared-subject comma list.
		"When this creature enters, draw a card.",
		// A self-contained ", or " disjunction is expandDisjunctiveTrigger's job.
		"Whenever another creature dies, or a creature card leaves your graveyard, draw a card.",
	}
	for _, source := range unchanged {
		if got := expandSharedSubjectTriggerUnion(source); got != source {
			t.Fatalf("expandSharedSubjectTriggerUnion(%q) = %q, want unchanged", source, got)
		}
	}
}
