package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImpostorOfTheSixthPride is the card definition for Impostor of the Sixth Pride.
//
// Type: Creature — Shapeshifter
// Cost: {1}{W}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
var ImpostorOfTheSixthPride = newImpostorOfTheSixthPride

func newImpostorOfTheSixthPride() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Impostor of the Sixth Pride",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
		`,
		},
	}
}
