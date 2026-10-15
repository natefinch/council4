package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArdentPlea is the card definition for Ardent Plea.
//
// Type: Enchantment
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
var ArdentPlea = newArdentPlea

func newArdentPlea() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Ardent Plea",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.ExaltedStaticBody,
				game.CascadeStaticBody,
			},
			OracleText: `
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
			Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
		`,
		},
	}
}
