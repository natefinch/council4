package game

import "testing"

// TestValidateCardDefAcceptsEnchantedPlayerFirstUpkeepPattern proves the Paradox
// Haze trigger pattern — a beginning-of-upkeep step scoped to the source's
// enchanted player and gated on the first upkeep each turn — passes validation.
func TestValidateCardDefAcceptsEnchantedPlayerFirstUpkeepPattern(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Valid Paradox Haze",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{Sequence: []Instruction{{Primitive: AddExtraUpkeepStep{}}}}.Ability(),
			Trigger: TriggerCondition{
				Type: TriggerAt,
				Pattern: TriggerPattern{
					Event:                             EventBeginningOfStep,
					Step:                              StepUpkeep,
					StepPlayerIsSourceEnchantedPlayer: true,
					FirstUpkeepStepEachTurn:           true,
				},
			},
		}},
	}}

	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

// TestValidateCardDefRejectsInvalidUpkeepStepPatternFields proves the new
// additional-upkeep-step trigger filters fail validation when applied to the
// wrong event or step.
func TestValidateCardDefRejectsInvalidUpkeepStepPatternFields(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
	}{
		{
			name: "enchanted-player step filter on non-step event",
			pattern: TriggerPattern{
				Event:                             EventAttackerDeclared,
				StepPlayerIsSourceEnchantedPlayer: true,
			},
		},
		{
			name: "first-upkeep filter on non-upkeep step",
			pattern: TriggerPattern{
				Event:                   EventBeginningOfStep,
				Step:                    StepDraw,
				FirstUpkeepStepEachTurn: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Invalid Upkeep Step Pattern",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: tt.pattern},
				}},
			}}

			issues := ValidateCardDef(card)

			if len(issues) != 1 || issues[0].Code != CardDefIssueInvalidSelection {
				t.Fatalf("issues = %+v, want one %s", issues, CardDefIssueInvalidSelection)
			}
		})
	}
}
