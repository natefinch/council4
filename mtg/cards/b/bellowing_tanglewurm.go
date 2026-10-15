package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BellowingTanglewurm is the card definition for Bellowing Tanglewurm.
//
// Type: Creature — Wurm
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	Other green creatures you control have intimidate.
var BellowingTanglewurm = newBellowingTanglewurm

func newBellowingTanglewurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bellowing Tanglewurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.IntimidateStaticBody,
				game.StaticAbility{
					Text: "Other green creatures you control have intimidate.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Intimidate},
						},
					},
				},
			},
			OracleText: `
			Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
			Other green creatures you control have intimidate.
		`,
		},
	}
}
