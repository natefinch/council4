package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrystallineSliver is the card definition for Crystalline Sliver.
//
// Type: Creature — Sliver
// Cost: {W}{U}
//
// Oracle text:
//
//	All Slivers have shroud. (They can't be the targets of spells or abilities.)
var CrystallineSliver = newCrystallineSliver

func newCrystallineSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Crystalline Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All Slivers have shroud. (They can't be the targets of spells or abilities.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			All Slivers have shroud. (They can't be the targets of spells or abilities.)
		`,
		},
	}
}
