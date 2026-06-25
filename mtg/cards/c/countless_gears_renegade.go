package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CountlessGearsRenegade is the card definition for Countless Gears Renegade.
//
// Type: Creature — Dwarf Artificer
// Cost: {1}{W}
//
// Oracle text:
//
//	Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, create a 1/1 colorless Servo artifact creature token.
var CountlessGearsRenegade = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name: "Countless Gears Renegade",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.W,
		}),
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf, types.Artificer},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
					InterveningIf: "if a permanent left the battlefield under your control this turn",
					InterveningCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Controller:    game.TriggerControllerYou,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						}, Window: game.EventHistoryCurrentTurn}),
					}),
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.CreateToken{
								Amount: game.Fixed(1),
								Source: game.TokenDef(countlessGearsRenegadeToken),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, create a 1/1 colorless Servo artifact creature token.
		`,
	},
}

var countlessGearsRenegadeToken = newCountlessGearsRenegadeToken()

func newCountlessGearsRenegadeToken() *game.CardDef {
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
