package game

import "testing"

// TestValidateCardDefBlockerScope requires a blocker-subject blocker-declared
// trigger to choose "blocks" (CR 509.3a) or "blocks a creature" (CR 509.3b),
// and rejects a scope where the blocker is not the subject.
func TestValidateCardDefBlockerScope(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
		valid   bool
	}{
		{"blocks unspecified", TriggerPattern{Event: EventBlockerDeclared, Source: TriggerSourceSelf}, false},
		{"blocks once", TriggerPattern{Event: EventBlockerDeclared, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerOnce}, true},
		{"blocks a creature", TriggerPattern{Event: EventBlockerDeclared, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerEachBlockedCreature}, true},
		{"attacks or blocks unspecified", TriggerPattern{Event: EventAttackerDeclared, UnionEvent: EventBlockerDeclared, Source: TriggerSourceSelf}, false},
		{"attacks or blocks once", TriggerPattern{Event: EventAttackerDeclared, UnionEvent: EventBlockerDeclared, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerOnce}, true},
		{"blocks or becomes blocked unspecified", TriggerPattern{Event: EventBlockerDeclared, UnionEvent: EventAttackerBecameBlocked, Source: TriggerSourceSelf}, false},
		{"becomes blocked by a creature", TriggerPattern{Event: EventBlockerDeclared, Subject: TriggerSubjectBlockedAttacker, Source: TriggerSourceSelf}, true},
		{"becomes blocked by a creature with scope", TriggerPattern{Event: EventBlockerDeclared, Subject: TriggerSubjectBlockedAttacker, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerOnce}, false},
		{"attacks with scope", TriggerPattern{Event: EventAttackerDeclared, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerOnce}, false},
		{"unknown scope", TriggerPattern{Event: EventBlockerDeclared, Source: TriggerSourceSelf, BlockerScope: BlockerTriggerEachBlockedCreature + 1}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Blocker Scope",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: test.pattern},
				}},
			}}
			issues := ValidateCardDef(card)
			if test.valid && len(issues) != 0 {
				t.Fatalf("issues = %+v, want none", issues)
			}
			if !test.valid && (len(issues) != 1 || issues[0].Code != CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want one %s", issues, CardDefIssueInvalidSelection)
			}
		})
	}
}
