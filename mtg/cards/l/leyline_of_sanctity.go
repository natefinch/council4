package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineOfSanctity is the card definition for Leyline of Sanctity.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	You have hexproof. (You can't be the target of spells or abilities your opponents control.)
var LeylineOfSanctity = newLeylineOfSanctity

func newLeylineOfSanctity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leyline of Sanctity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{},
				game.PlayerHexproofStaticBody,
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			You have hexproof. (You can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
