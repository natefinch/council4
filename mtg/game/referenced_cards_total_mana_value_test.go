package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestValidateReferencedCardsTotalManaValueLink(t *testing.T) {
	const key LinkedKey = "milled"
	sequence := []Instruction{
		{Primitive: MoveTopOfLibrary{Amount: Fixed(3), Player: ControllerReference(), Destination: zone.Graveyard, PublishLinked: key}},
		{Primitive: Damage{
			Amount: Dynamic(DynamicAmount{
				Kind:      DynamicAmountReferencedCardsTotalManaValue,
				LinkedKey: key,
			}),
			Recipient: PlayerDamageRecipient(TargetPlayerReference(0)),
		}},
	}
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPlayer}}
	if err := ValidateInstructionSequence(sequence, targets); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}
	sequence[1].Primitive = Damage{
		Amount:    Dynamic(DynamicAmount{Kind: DynamicAmountReferencedCardsTotalManaValue}),
		Recipient: PlayerDamageRecipient(TargetPlayerReference(0)),
	}
	if err := ValidateInstructionSequence(sequence, targets); err == nil {
		t.Fatal("ValidateInstructionSequence() accepted a missing linked key")
	}
}
