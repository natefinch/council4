package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MirrorEntity is the card definition for Mirror Entity.
//
// Type: Creature — Shapeshifter
// Cost: {2}{W}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.
var MirrorEntity = newMirrorEntity

func newMirrorEntity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mirror Entity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:    game.LayerPowerToughnessSet,
													SetPowerDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountX,
														CounterKind: counter.PlusOnePlusOne,
													}),
													SetToughnessDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountX,
														CounterKind: counter.PlusOnePlusOne,
													}),
												},
												game.ContinuousEffect{
													Duration:             game.DurationPermanent,
													Group:                game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:                game.LayerType,
													AddEveryCreatureType: true,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Changeling (This card is every creature type.)
			{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.
		`,
		},
	}
}
