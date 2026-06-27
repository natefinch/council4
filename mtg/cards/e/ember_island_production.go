package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmberIslandProduction is the card definition for Ember Island Production.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Choose one —
//	• Create a token that's a copy of target creature you control, except it's not legendary and it's a 4/4 Hero in addition to its other types.
//	• Create a token that's a copy of target creature an opponent controls, except it's not legendary and it's a 2/2 Coward in addition to its other types.
var EmberIslandProduction = newEmberIslandProduction()

func newEmberIslandProduction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ember Island Production",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Create a token that's a copy of target creature you control, except it's not legendary and it's a 4/4 Hero in addition to its other types.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:          game.TokenCopySourceObject,
										Object:          game.TargetPermanentReference(0),
										SetNotLegendary: true,
										SetPower:        opt.Val(game.PT{Value: 4}),
										SetToughness:    opt.Val(game.PT{Value: 4}),
										AddSubtypes:     []types.Sub{types.Hero},
									}),
								},
							},
						},
					},
					game.Mode{
						Text: "Create a token that's a copy of target creature an opponent controls, except it's not legendary and it's a 2/2 Coward in addition to its other types.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:          game.TokenCopySourceObject,
										Object:          game.TargetPermanentReference(0),
										SetNotLegendary: true,
										SetPower:        opt.Val(game.PT{Value: 2}),
										SetToughness:    opt.Val(game.PT{Value: 2}),
										AddSubtypes:     []types.Sub{types.Coward},
									}),
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
			• Create a token that's a copy of target creature you control, except it's not legendary and it's a 4/4 Hero in addition to its other types.
			• Create a token that's a copy of target creature an opponent controls, except it's not legendary and it's a 2/2 Coward in addition to its other types.
		`,
		},
	}
}
