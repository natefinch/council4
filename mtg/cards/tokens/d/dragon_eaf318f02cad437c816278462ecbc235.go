package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dragon
//
// Type: Token Creature — Dragon
//
// Oracle text:
//   Flying

// DragonTokeneaf318f02cad437c816278462ecbc235 is the card definition for Dragon.
var DragonTokeneaf318f02cad437c816278462ecbc235 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Dragon",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
