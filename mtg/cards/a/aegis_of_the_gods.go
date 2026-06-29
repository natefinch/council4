package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AegisOfTheGods is the card definition for Aegis of the Gods.
//
// Type: Enchantment Creature — Human Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	You have hexproof. (You can't be the target of spells or abilities your opponents control.)
var AegisOfTheGods = newAegisOfTheGods()

func newAegisOfTheGods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Aegis of the Gods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.PlayerHexproofStaticBody,
			},
			OracleText: `
			You have hexproof. (You can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
