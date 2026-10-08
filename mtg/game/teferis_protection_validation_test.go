package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

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
		MoveResolvingSpell{Destination: zone.Exile},
		// Resolving spells may also return their exact stack card to hand.
		MoveResolvingSpell{Destination: zone.Hand},
	}
	for _, primitive := range valid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err != nil {
			t.Fatalf("%T validation failed: %v", primitive, err)
		}
	}
	invalid := []Primitive{
		ApplyRule{RuleEffects: []RuleEffect{{Kind: RuleEffectPlayerProtection}}},
		PhaseOut{},
		// A resolving-spell move used to be invalid when it named an object;
		// the object field is gone, so the equivalent malformed instruction is
		// one whose destination the runtime cannot reach.
		MoveResolvingSpell{},
		MoveResolvingSpell{Destination: zone.Battlefield},
	}
	for _, primitive := range invalid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}); err == nil {
			t.Fatalf("%T validation succeeded, want failure", primitive)
		}
	}
}
