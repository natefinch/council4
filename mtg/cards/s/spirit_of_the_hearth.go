package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpiritOfTheHearth is the card definition for Spirit of the Hearth.
//
// Type: Creature — Cat Spirit
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Flying
//	You have hexproof. (You can't be the target of spells or abilities your opponents control.)
var SpiritOfTheHearth = newSpiritOfTheHearth

func newSpiritOfTheHearth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spirit of the Hearth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Spirit},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.PlayerHexproofStaticBody,
			},
			OracleText: `
			Flying
			You have hexproof. (You can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
