package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DriftOfPhantasms is the card definition for Drift of Phantasms.
//
// Type: Creature — Spirit
// Cost: {2}{U}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Flying
//	Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var DriftOfPhantasms = newDriftOfPhantasms

func newDriftOfPhantasms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Drift of Phantasms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.U}, 3),
			},
			OracleText: `
			Defender (This creature can't attack.)
			Flying
			Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
