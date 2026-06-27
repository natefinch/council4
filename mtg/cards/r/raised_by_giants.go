package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaisedByGiants is the card definition for Raised by Giants.
//
// Type: Legendary Enchantment — Background
// Cost: {5}{G}
//
// Oracle text:
//
//	Commander creatures you own have base power and toughness 10/10 and are Giants in addition to their other types.
var RaisedByGiants = newRaisedByGiants()

func newRaisedByGiants() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Raised by Giants",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							SetPower:     opt.Val(game.PT{Value: 10}),
							SetToughness: opt.Val(game.PT{Value: 10}),
						},
						game.ContinuousEffect{
							Layer:       game.LayerType,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							AddSubtypes: []types.Sub{types.Sub("Giant")},
						},
					},
				},
			},
			OracleText: `
			Commander creatures you own have base power and toughness 10/10 and are Giants in addition to their other types.
		`,
		},
	}
}
