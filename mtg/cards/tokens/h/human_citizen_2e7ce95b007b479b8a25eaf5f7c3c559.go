package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Citizen
//
// Type: Token Creature — Human Citizen
//
// Oracle text:

// HumanCitizenToken2e7ce95b007b479b8a25eaf5f7c3c559 is the card definition for Human Citizen.
var HumanCitizenToken2e7ce95b007b479b8a25eaf5f7c3c559 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Human Citizen",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Citizen},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
