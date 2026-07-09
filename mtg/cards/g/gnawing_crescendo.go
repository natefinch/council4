package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GnawingCrescendo is the card definition for Gnawing Crescendo.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Creatures you control get +2/+0 until end of turn. Whenever a nontoken creature you control dies this turn, create a 1/1 black Rat creature token with "This token can't block."
var GnawingCrescendo = newGnawingCrescendo

func newGnawingCrescendo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gnawing Crescendo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:      game.LayerPowerToughnessModify,
									Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								EventPattern: opt.Val(game.TriggerPattern{
									Event:            game.EventPermanentDied,
									Controller:       game.TriggerControllerYou,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
								}),
								Window: game.DelayedWindowThisTurn,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.CreateToken{
												Amount: game.Fixed(1),
												Source: game.TokenDef(gnawingCrescendoToken),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +2/+0 until end of turn. Whenever a nontoken creature you control dies this turn, create a 1/1 black Rat creature token with "This token can't block."
		`,
		},
	}
}

var gnawingCrescendoToken = newGnawingCrescendoToken()

func newGnawingCrescendoToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rat",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
						},
					},
				},
			},
		},
	}
}
