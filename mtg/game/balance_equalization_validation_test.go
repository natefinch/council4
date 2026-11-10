package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestValidateBalanceEqualizationShapes(t *testing.T) {
	lands := Selection{RequiredTypes: []types.Card{types.Land}}
	fewestLands := Dynamic(DynamicAmount{Kind: DynamicAmountFewestPermanentsControlled, Selection: &lands})
	fewestHand := Dynamic(DynamicAmount{Kind: DynamicAmountFewestCardsInHand})
	valid := []Instruction{
		{Primitive: KeepPermanents{PlayerGroup: AllPlayersReference(), AffectedSelection: lands, Keep: fewestLands}},
		{Primitive: Discard{PlayerGroup: AllPlayersReference(), KeepInHand: true, Amount: fewestHand}},
	}
	if err := ValidateInstructionSequence(valid, nil); err != nil {
		t.Fatalf("ValidateInstructionSequence() error = %v", err)
	}

	yourLands := Selection{RequiredTypes: []types.Card{types.Land}, Controller: ControllerYou}
	invalid := map[string]Primitive{
		"controller-restricted fewest count": KeepPermanents{
			PlayerGroup: AllPlayersReference(),
			Keep:        Dynamic(DynamicAmount{Kind: DynamicAmountFewestPermanentsControlled, Selection: &yourLands}),
		},
		"keep-in-hand entire hand":  Discard{PlayerGroup: AllPlayersReference(), KeepInHand: true, EntireHand: true},
		"keep-in-hand at random":    Discard{PlayerGroup: AllPlayersReference(), KeepInHand: true, AtRandom: true, Amount: fewestHand},
		"keep-in-hand publish link": Discard{Player: ControllerReference(), KeepInHand: true, PublishLinked: "kept", Amount: fewestHand},
	}
	for name, primitive := range invalid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: primitive}}, nil); err == nil {
			t.Errorf("%s: ValidateInstructionSequence() accepted an invalid shape", name)
		}
	}
}
