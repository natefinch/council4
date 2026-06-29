package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrueBeliever is the card definition for True Believer.
//
// Type: Creature — Human Cleric
// Cost: {W}{W}
//
// Oracle text:
//
//	You have shroud. (You can't be the target of spells or abilities.)
var TrueBeliever = newTrueBeliever()

func newTrueBeliever() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "True Believer",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.PlayerShroudStaticBody,
			},
			OracleText: `
			You have shroud. (You can't be the target of spells or abilities.)
		`,
		},
	}
}
