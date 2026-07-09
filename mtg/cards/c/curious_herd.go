package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CuriousHerd is the card definition for Curious Herd.
//
// Type: Instant
// Cost: {3}{G}
//
// Oracle text:
//
//	Choose target opponent. You create X 3/3 green Beast creature tokens, where X is the number of artifacts that player controls.
var CuriousHerd = newCuriousHerd

func newCuriousHerd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Curious Herd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							}),
							Source: game.TokenDef(curiousHerdToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose target opponent. You create X 3/3 green Beast creature tokens, where X is the number of artifacts that player controls.
		`,
		},
	}
}

var curiousHerdToken = newCuriousHerdToken()

func newCuriousHerdToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
