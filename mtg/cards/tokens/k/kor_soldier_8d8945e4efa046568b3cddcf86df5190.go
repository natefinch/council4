package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kor Soldier
//
// Type: Token Creature — Kor Soldier
//
// Oracle text:

// KorSoldierToken8d8945e4efa046568b3cddcf86df5190 is the card definition for Kor Soldier.
var KorSoldierToken8d8945e4efa046568b3cddcf86df5190 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Kor Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kor, types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
