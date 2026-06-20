package game

import "testing"

func TestValidateTeferisProtectionPrimitives(t *testing.T) {
	t.Parallel()
	valid := []Primitive{
		ApplyRule{
			RuleEffects: []RuleEffect{{
				Kind:           RuleEffectLifeTotalCantChange,
				AffectedPlayer: PlayerYou,
			}},
			Duration: DurationUntilYourNextTurn,
		},
		PhaseOut{Group: BattlefieldGroup(Selection{Controller: ControllerYou})},
		Exile{SourceSpell: true},
	}
	for _, primitive := range valid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err != nil {
			t.Fatalf("%T validation failed: %v", primitive, err)
		}
	}
	invalid := []Primitive{
		ApplyRule{RuleEffects: []RuleEffect{{Kind: RuleEffectPlayerProtection}}},
		PhaseOut{},
		Exile{SourceSpell: true, Object: SourcePermanentReference()},
	}
	for _, primitive := range invalid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err == nil {
			t.Fatalf("%T validation succeeded, want failure", primitive)
		}
	}
}
