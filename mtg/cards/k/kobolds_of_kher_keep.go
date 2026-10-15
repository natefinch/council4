package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KoboldsOfKherKeep is the card definition for Kobolds of Kher Keep.
//
// Type: Creature — Kobold
// Cost: {0}
//
// Oracle text:
var KoboldsOfKherKeep = newKoboldsOfKherKeep

func newKoboldsOfKherKeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kobolds of Kher Keep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kobold},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
