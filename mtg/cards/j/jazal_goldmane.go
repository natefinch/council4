package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JazalGoldmane is the card definition for Jazal Goldmane.
//
// Type: Legendary Creature — Cat Warrior
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	{3}{W}{W}: Attacking creatures you control get +X/+X until end of turn, where X is the number of attacking creatures.
var JazalGoldmane = newJazalGoldmane

func newJazalGoldmane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Jazal Goldmane",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{W}{W}: Attacking creatures you control get +X/+X until end of turn, where X is the number of attacking creatures.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.W,
						cost.W,
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
													Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
													Layer:    game.LayerPowerToughnessModify,
													PowerDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountSelector,
														Multiplier:  1,
														CounterKind: counter.PlusOnePlusOne,
														Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
													}),
													ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountSelector,
														Multiplier:  1,
														CounterKind: counter.PlusOnePlusOne,
														Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
													}),
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
			First strike (This creature deals combat damage before creatures without first strike.)
			{3}{W}{W}: Attacking creatures you control get +X/+X until end of turn, where X is the number of attacking creatures.
		`,
		},
	}
}
