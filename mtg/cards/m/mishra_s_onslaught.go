package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MishraSOnslaught is the card definition for Mishra's Onslaught.
//
// Type: Instant
// Cost: {3}{R}
//
// Oracle text:
//
//	Choose one —
//	• Create two 1/1 colorless Soldier artifact creature tokens.
//	• Creatures you control get +2/+0 until end of turn.
var MishraSOnslaught = newMishraSOnslaught

func newMishraSOnslaught() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mishra's Onslaught",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Create two 1/1 colorless Soldier artifact creature tokens.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(mishraSOnslaughtToken),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control get +2/+0 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:   game.DurationPermanent,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 2,
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
			}),
			OracleText: `
			Choose one —
			• Create two 1/1 colorless Soldier artifact creature tokens.
			• Creatures you control get +2/+0 until end of turn.
		`,
		},
	}
}

var mishraSOnslaughtToken = newMishraSOnslaughtToken()

func newMishraSOnslaughtToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
