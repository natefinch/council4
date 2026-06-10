package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dragon Elemental
//
// Type: Token Creature — Dragon Elemental
//
// Oracle text:
//   Flying
//   Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)

// DragonElementalToken396a7698cfbf4655b0d0b1265f4093a9 is the card definition for Dragon Elemental.
var DragonElementalToken396a7698cfbf4655b0d0b1265f4093a9 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dragon Elemental",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon, types.Elemental},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.ProwessStaticBody,
		},
		OracleText: `
			Flying
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
	},
}
