package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tah-Crop Skirmisher
//
// Type: Token Creature — Zombie Naga Warrior
//
// Oracle text:

// TahCropSkirmisherTokenf1444785ae10420d9621be0df408d480 is the card definition for Tah-Crop Skirmisher.
var TahCropSkirmisherTokenf1444785ae10420d9621be0df408d480 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Tah-Crop Skirmisher",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Sub("Naga"), types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
