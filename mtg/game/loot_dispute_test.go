package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestLootDisputeTriggerVocabularyValidates(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name: "Reusable Dungeon Trigger",
		TriggeredAbilities: []TriggeredAbility{{
			Trigger: TriggerCondition{Type: TriggerWhenever, Pattern: TriggerPattern{
				Event:  EventCompletedDungeon,
				Player: TriggerPlayerYou,
			}},
			Content: Mode{Sequence: []Instruction{{Primitive: CreateToken{
				Amount:    Fixed(1),
				Source:    TokenDef(&CardDef{CardFace: CardFace{Name: "Dragon"}}),
				Recipient: opt.Val(ControllerReference()),
			}}}}.Ability(),
		}, {
			Trigger: TriggerCondition{Type: TriggerWhenever, Pattern: TriggerPattern{
				Event:                    EventAttackerDeclared,
				Controller:               TriggerControllerYou,
				Player:                   TriggerPlayerInitiative,
				AttackRecipient:          AttackRecipientPlayer,
				OneOrMore:                true,
				OneOrMorePerAttackTarget: true,
			}},
			Content: Mode{Sequence: []Instruction{{Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenDef(&CardDef{CardFace: CardFace{Name: "Treasure"}}),
			}}}}.Ability(),
		}},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %#v, want none", issues)
	}
}
