package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AirdropAeronauts is the card definition for Airdrop Aeronauts.
//
// Type: Creature — Dwarf Scout
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, you gain 5 life.
var AirdropAeronauts = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name: "Airdrop Aeronauts",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.W,
			cost.W,
		}),
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf, types.Scout},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
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
							Primitive: game.GainLife{
								Amount: game.Fixed(5),
								Player: game.ControllerReference(),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Flying
			Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, you gain 5 life.
		`,
	},
}
