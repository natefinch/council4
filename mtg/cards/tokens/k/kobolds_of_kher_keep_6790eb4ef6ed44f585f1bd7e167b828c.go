package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kobolds of Kher Keep
//
// Type: Token Creature — Kobold
//
// Oracle text:

// KoboldsOfKherKeepToken6790eb4ef6ed44f585f1bd7e167b828c is the card definition for Kobolds of Kher Keep.
var KoboldsOfKherKeepToken6790eb4ef6ed44f585f1bd7e167b828c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Kobolds of Kher Keep",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kobold},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
