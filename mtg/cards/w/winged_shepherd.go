package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WingedShepherd is the card definition for Winged Shepherd.
//
// Type: Creature — Angel
// Cost: {5}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	Cycling {W} ({W}, Discard this card: Draw a card.)
var WingedShepherd = newWingedShepherd

func newWingedShepherd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Winged Shepherd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.W}),
			},
			OracleText: `
			Flying, vigilance
			Cycling {W} ({W}, Discard this card: Draw a card.)
		`,
		},
	}
}
