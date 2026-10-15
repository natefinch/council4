package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhoxPikemaster is the card definition for Rhox Pikemaster.
//
// Type: Creature — Rhino Soldier
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Other Soldier creatures you control have first strike.
var RhoxPikemaster = newRhoxPikemaster

func newRhoxPikemaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rhox Pikemaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					Text: "Other Soldier creatures you control have first strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Soldier")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			Other Soldier creatures you control have first strike.
		`,
		},
	}
}
