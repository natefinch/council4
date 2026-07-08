package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VanguardOfTheRestless is the card definition for Vanguard of the Restless.
//
// Type: Creature — Spirit Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Spirits you control get +1/+1 for each time you've cast your commander from the command zone this game.
//	Whenever a Spirit you control enters, you may pay {2}{W}. If you do, return this card from your graveyard to the battlefield.
var VanguardOfTheRestless = newVanguardOfTheRestless

func newVanguardOfTheRestless() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Vanguard of the Restless",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit")}}),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCommanderCastCount,
								Multiplier: 1,
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCommanderCastCount,
								Multiplier: 1,
							}),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {2}{W}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
											cost.W,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Spirits you control get +1/+1 for each time you've cast your commander from the command zone this game.
			Whenever a Spirit you control enters, you may pay {2}{W}. If you do, return this card from your graveyard to the battlefield.
		`,
		},
	}
}
