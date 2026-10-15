package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MadrushCyclops is the card definition for Madrush Cyclops.
//
// Type: Creature — Cyclops Warrior
// Cost: {1}{B}{R}{G}
//
// Oracle text:
//
//	Creatures you control have haste.
var MadrushCyclops = newMadrushCyclops

func newMadrushCyclops() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Madrush Cyclops",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cyclops, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have haste.
		`,
		},
	}
}
