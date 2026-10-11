package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestActualCardQuantityValidation(t *testing.T) {
	selection := Selection{RequiredTypes: []types.Card{types.Creature}}
	quantity := Dynamic(DynamicAmount{Kind: DynamicAmountPreviousEffectResult, ResultKey: "cards", Selection: &selection})
	sequence := []Instruction{
		{Primitive: MoveTopOfLibrary{Player: ControllerReference(), Amount: Fixed(2), Destination: zone.Graveyard}, PublishResult: "cards"},
		{Primitive: GainLife{Player: ControllerReference(), Amount: quantity},
			ResultGate: opt.Val(InstructionResultGate{Key: "cards", Accepted: TriTrue})},
	}
	if err := ValidateInstructionSequence(sequence); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]Instruction){
		func(s []Instruction) { s[0].Primitive = Destroy{Object: SourcePermanentReference()} },
		func(s []Instruction) { s[0].PublishResult = "another" },
		func(s []Instruction) { s[1].ResultGate = opt.V[InstructionResultGate]{} },
		func(s []Instruction) { s[1].ResultGate.Val.Accepted = TriAny },
		func(s []Instruction) {
			s[0].Optional, s[0].TemptingOffer = true, true
			s[0].OptionalActorGroup = opt.Val(OpponentsReference())
			s[0].Primitive = MoveTopOfLibrary{Player: GroupOfferMemberReference(), Amount: Fixed(2), Destination: zone.Graveyard}
		},
		func(s []Instruction) {
			s[0].Primitive = MoveTopOfLibrary{Player: ControllerReference(), Amount: Fixed(2), Destination: zone.Exile, FaceDown: true}
		},
		func(s []Instruction) {
			s[0].PublishResult = ""
			s[0].Primitive = MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Hand,
				PublishDepartureCharacteristics: LibraryCardCharacteristics{ManaValue: "cards"}}
		},
	} {
		broken := append([]Instruction(nil), sequence...)
		mutate(broken)
		if err := ValidateInstructionSequence(broken); err == nil {
			t.Fatal("invalid snapshot domain or availability gate accepted")
		}
	}
	broken := append([]Instruction(nil), sequence...)
	contextual := Selection{Controller: ControllerYou}
	broken[1].Primitive = GainLife{Player: ControllerReference(), Amount: Dynamic(DynamicAmount{
		Kind: DynamicAmountPreviousEffectResult, ResultKey: "cards", Selection: &contextual,
	})}
	if err := ValidateInstructionSequence(broken); err == nil {
		t.Fatal("contextual/live predicate accepted as a frozen card filter")
	}
}
