package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TideDrifter is the card definition for Tide Drifter.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Other colorless creatures you control get +0/+1.
var TideDrifter = newTideDrifter

func newTideDrifter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tide Drifter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.StaticAbility{
					Text: "Other colorless creatures you control get +0/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			Other colorless creatures you control get +0/+1.
		`,
		},
	}
}
