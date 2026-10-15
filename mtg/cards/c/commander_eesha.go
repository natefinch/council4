package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommanderEesha is the card definition for Commander Eesha.
//
// Type: Legendary Creature — Bird Soldier
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying, protection from creatures
var CommanderEesha = newCommanderEesha

func newCommanderEesha() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Commander Eesha",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bird, types.Soldier},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProtectionFromTypesStaticAbility(types.Creature),
			},
			OracleText: `
			Flying, protection from creatures
		`,
		},
	}
}
