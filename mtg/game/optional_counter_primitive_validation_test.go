package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestValidateOptionalCounterForEachPlayerLinkedGoad(t *testing.T) {
	const key LinkedKey = "optional-counter"
	err := ValidateInstructionSequence([]Instruction{
		{Primitive: OptionalCounterForEachPlayer{
			Players:       AllPlayersReference(),
			Selection:     Selection{RequiredTypes: []types.Card{types.Creature}},
			Amount:        Fixed(2),
			CounterKind:   counter.PlusOnePlusOne,
			PublishLinked: key,
		}},
		{Primitive: Goad{
			Group:         LinkedObjectsGroup(key),
			ConsumeLinked: true,
		}},
	})
	if err != nil {
		t.Fatalf("ValidateInstructionSequence() = %v", err)
	}
}

func TestValidateOptionalCounterForEachPlayerRequiresLink(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: OptionalCounterForEachPlayer{
		Players:     AllPlayersReference(),
		Selection:   Selection{RequiredTypes: []types.Card{types.Creature}},
		Amount:      Fixed(2),
		CounterKind: counter.PlusOnePlusOne,
	}}})
	if err == nil {
		t.Fatal("ValidateInstructionSequence() accepted missing PublishLinked")
	}
}

func TestValidateConsumeLinkedGoadRequiresLinkedGroup(t *testing.T) {
	err := ValidateInstructionSequence([]Instruction{{Primitive: Goad{
		Group:         BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Creature}}),
		ConsumeLinked: true,
	}}})
	if err == nil {
		t.Fatal("ValidateInstructionSequence() accepted consume-linked battlefield goad")
	}
}
