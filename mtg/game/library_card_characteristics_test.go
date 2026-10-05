package game

import "testing"

func TestLibraryCardCharacteristicOutputsValidateExactDomainAndSeparateKeys(t *testing.T) {
	for _, tc := range []struct {
		name      string
		primitive Primitive
		valid     bool
	}{
		{"singular look", LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "power", Toughness: "toughness"}}, true},
		{"singular reveal", Reveal{Player: ControllerReference(), Amount: Fixed(1), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "power"}}, true},
		{"multi-card", Reveal{Player: ControllerReference(), Amount: Fixed(2), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "power"}}, false},
		{"dynamic amount", Reveal{Player: ControllerReference(), Amount: Dynamic(DynamicAmount{Kind: DynamicAmountX}), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "power"}}, false},
		{"alias characteristics", LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "scalar", Toughness: "scalar"}}, false},
		{"alias object", LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "card",
			PublishCharacteristics: LibraryCardCharacteristics{Power: "card"}}, false},
		{"no identity", Reveal{Player: ControllerReference(), Amount: Fixed(1),
			PublishCharacteristics: LibraryCardCharacteristics{Power: "power"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			instruction := Instruction{Primitive: tc.primitive,
				LocalProducts: LocalProducts{Results: PublishedScalarKeys(tc.primitive), Links: []LinkedKey{"card"}}}
			err := ValidateInstructionSequence([]Instruction{instruction})
			if (err == nil) != tc.valid {
				t.Fatalf("validation=%v, want valid=%t", err, tc.valid)
			}
		})
	}
}

func TestLibraryCardCharacteristicConsumerRequiresItsExactScalarProducer(t *testing.T) {
	producer := Instruction{Primitive: LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "card",
		PublishCharacteristics: LibraryCardCharacteristics{Power: "power"}}}
	consumer := Instruction{Primitive: GainLife{Player: ControllerReference(),
		Amount: Dynamic(DynamicAmount{Kind: DynamicAmountPreviousEffectResult, ResultKey: "power"})}}
	if err := ValidateInstructionSequence([]Instruction{producer, consumer}); err != nil {
		t.Fatal(err)
	}
	consumer.Primitive = GainLife{Player: ControllerReference(),
		Amount: Dynamic(DynamicAmount{Kind: DynamicAmountPreviousEffectResult, ResultKey: "toughness"})}
	if err := ValidateInstructionSequence([]Instruction{producer, consumer}); err == nil {
		t.Fatal("undeclared characteristic consumer passed validation")
	}
}
