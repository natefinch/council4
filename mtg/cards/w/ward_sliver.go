package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WardSliver is the card definition for Ward Sliver.
//
// Type: Creature — Sliver
// Cost: {4}{W}
//
// Oracle text:
//
//	As this creature enters, choose a color.
//	All Slivers have protection from the chosen color.
var WardSliver = newWardSliver

func newWardSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ward Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromChosenColorStaticAbility()),
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this creature enters, choose a color."),
			},
			OracleText: `
			As this creature enters, choose a color.
			All Slivers have protection from the chosen color.
		`,
		},
	}
}
