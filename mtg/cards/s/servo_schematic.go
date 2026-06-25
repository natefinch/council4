package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ServoSchematic is the card definition for Servo Schematic.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	When this artifact enters or is put into a graveyard from the battlefield, create a 1/1 colorless Servo artifact creature token.
var ServoSchematic = newServoSchematic()

func newServoSchematic() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Servo Schematic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventPermanentDied,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(servoSchematicToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters or is put into a graveyard from the battlefield, create a 1/1 colorless Servo artifact creature token.
		`,
		},
	}
}

var servoSchematicToken = newServoSchematicToken()

func newServoSchematicToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Servo",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Servo},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
