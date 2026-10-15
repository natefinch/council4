package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiftburstHellion is the card definition for Riftburst Hellion.
//
// Type: Creature — Hellion
// Cost: {5}{R}{G}
//
// Oracle text:
//
//	Reach
//	Disguise {4}{R/G}{R/G} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
var RiftburstHellion = newRiftburstHellion

func newRiftburstHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Riftburst Hellion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					Text: "Disguise {4}{R/G}{R/G}",
					KeywordAbilities: []game.KeywordAbility{
						game.DisguiseKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.HybridMana(mana.R, mana.G),
								cost.HybridMana(mana.R, mana.G),
							},
						},
					},
				},
			},
			OracleText: `
			Reach
			Disguise {4}{R/G}{R/G} (You may cast this card face down for {3} as a 2/2 creature with ward {2}. Turn it face up any time for its disguise cost.)
		`,
		},
	}
}
