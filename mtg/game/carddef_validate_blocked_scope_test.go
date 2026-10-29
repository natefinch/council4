package game

import "testing"

// TestValidateCardDefBlockedScope requires an attacker-subject
// attacker-became-blocked trigger to choose "becomes blocked" (CR 509.3c) or
// "becomes blocked by a creature" (CR 509.3d), and rejects a scope anywhere
// else.
func TestValidateCardDefBlockedScope(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
		valid   bool
	}{
		{"becomes blocked unspecified", TriggerPattern{Event: EventAttackerBecameBlocked, Source: TriggerSourceSelf}, false},
		{"becomes blocked once", TriggerPattern{Event: EventAttackerBecameBlocked, Source: TriggerSourceSelf, BlockedScope: BlockedTriggerOnce}, true},
		{"becomes blocked by a creature", TriggerPattern{Event: EventAttackerBecameBlocked, Source: TriggerSourceSelf, BlockedScope: BlockedTriggerEachBlocker}, true},
		{"blocks or becomes blocked by a creature", TriggerPattern{Event: EventBlockerDeclared, BlockerScope: BlockerTriggerEachBlockedCreature, UnionEvent: EventAttackerBecameBlocked, BlockedScope: BlockedTriggerEachBlocker, Source: TriggerSourceSelf}, true},
		{"blocks or becomes blocked unspecified", TriggerPattern{Event: EventBlockerDeclared, BlockerScope: BlockerTriggerOnce, UnionEvent: EventAttackerBecameBlocked, Source: TriggerSourceSelf}, false},
		{"blocks with scope", TriggerPattern{Event: EventBlockerDeclared, BlockerScope: BlockerTriggerOnce, Source: TriggerSourceSelf, BlockedScope: BlockedTriggerOnce}, false},
		{"blocked-attacker subject with scope", TriggerPattern{Event: EventBlockerDeclared, Subject: TriggerSubjectBlockedAttacker, Source: TriggerSourceSelf, BlockedScope: BlockedTriggerEachBlocker}, false},
		{"unknown scope", TriggerPattern{Event: EventAttackerBecameBlocked, Source: TriggerSourceSelf, BlockedScope: BlockedTriggerEachBlocker + 1}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Blocked Scope",
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
