package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BolracClanBasher is the card definition for Bolrac-Clan Basher.
//
// Type: Creature — Cyclops Warrior
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Double strike, trample
//	Disguise {3}{R}{R} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
var BolracClanBasher = newBolracClanBasher

func newBolracClanBasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Bolrac-Clan Basher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cyclops, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "Disguise {3}{R}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.DisguiseKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.R,
								cost.R,
							},
						},
					},
				},
			},
			OracleText: `
			Double strike, trample
			Disguise {3}{R}{R} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
		`,
		},
	}
}
