package parser

import (
	"testing"
)

// TestParseSourceStateConditionAttachment verifies the source-permanent-state
// condition recognizer accepts both the bare pronoun subject ("it's equipped",
// "it is enchanted") and the explicit "this creature is <state>" subject for
// attachment, tap, and combat states, binding the inspected permanent to the
// source.
func TestParseSourceStateConditionAttachment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		condition  string
		tapped     ConditionTappedState
		combat     ConditionCombatState
		attachment ConditionAttachmentState
	}{
		{
			name:       "pronoun equipped",
			condition:  "it's equipped",
			attachment: ConditionAttachmentEquipped,
		},
		{
			name:       "pronoun enchanted",
			condition:  "it is enchanted",
			attachment: ConditionAttachmentEnchanted,
		},
		{
			name:      "pronoun attacking",
			condition: "it's attacking",
			combat:    ConditionCombatAttacking,
		},
		{
			name:      "pronoun untapped",
			condition: "it's untapped",
			tapped:    ConditionTappedFalse,
		},
		{
			name:       "explicit subject equipped",
			condition:  "this creature is equipped",
			attachment: ConditionAttachmentEquipped,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clause := parseSingleConditionClause(t, test.condition)
			if clause.Predicate != ConditionPredicateObjectMatches {
				t.Fatalf("predicate = %q, want %q", clause.Predicate, ConditionPredicateObjectMatches)
			}
			if clause.ObjectBinding != ConditionObjectBindingSource {
				t.Fatalf("object binding = %q, want %q", clause.ObjectBinding, ConditionObjectBindingSource)
			}
			if clause.Selection.Attachment != test.attachment {
				t.Fatalf("attachment = %q, want %q", clause.Selection.Attachment, test.attachment)
			}
			if clause.Selection.Tapped != test.tapped {
				t.Fatalf("tapped = %q, want %q", clause.Selection.Tapped, test.tapped)
			}
			if clause.Selection.CombatState != test.combat {
				t.Fatalf("combat state = %q, want %q", clause.Selection.CombatState, test.combat)
			}
		})
	}
}
