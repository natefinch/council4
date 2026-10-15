package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Avatar is the card definition for Avatar.
//
// Type: Token Creature — Avatar
//
// Oracle text:
//   This creature's power and toughness are each equal to your life total.

// AvatarToken6a20eb3c4ebb45609b77da35a4c2f207 is the card definition for Avatar.
var AvatarToken6a20eb3c4ebb45609b77da35a4c2f207 = newAvatarToken6a20eb3c4ebb45609b77da35a4c2f207()

func newAvatarToken6a20eb3c4ebb45609b77da35a4c2f207() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:             "Avatar",
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Avatar},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}),
			OracleText: `
			This creature's power and toughness are each equal to your life total.
		`,
		},
	}
}
