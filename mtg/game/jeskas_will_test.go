package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestJeskasWillAbilityContentValidates(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:  "Jeska's Will",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(AbilityContent{
			Modes: []Mode{
				{Sequence: []Instruction{{Primitive: AddMana{Amount: Fixed(1), ManaColor: mana.R}}}},
				{Sequence: []Instruction{{Primitive: ImpulseExile{
					Player: ControllerReference(), Amount: Fixed(3), Duration: DurationThisTurn,
				}}}},
			},
			MinModes: 1,
			MaxModes: 1,
			ModeChoiceBonus: ModeChoiceBonus{
				Condition:          ModeChoiceConditionControlsCommander,
				AdditionalMaxModes: 1,
			},
		}),
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("ValidateCardDef() = %+v", issues)
	}
}

func TestJeskasWillTypedFieldsFailClosed(t *testing.T) {
	tests := []struct {
		name    string
		content AbilityContent
	}{
		{
			name: "bonus exceeds modes",
			content: AbilityContent{
				Modes:    []Mode{{Sequence: []Instruction{{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}}}}},
				MinModes: 1, MaxModes: 1,
				ModeChoiceBonus: ModeChoiceBonus{Condition: ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1},
			},
		},
		{
			name: "impulse wrong duration",
			content: Mode{Sequence: []Instruction{{Primitive: ImpulseExile{
				Player: ControllerReference(), Amount: Fixed(3), Duration: DurationPermanent,
			}}}}.Ability(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Invalid", Types: []types.Card{types.Sorcery}, SpellAbility: opt.Val(test.content),
			}}
			if issues := ValidateCardDef(card); len(issues) == 0 {
				t.Fatal("ValidateCardDef() accepted invalid typed fields")
			}
		})
	}
}
