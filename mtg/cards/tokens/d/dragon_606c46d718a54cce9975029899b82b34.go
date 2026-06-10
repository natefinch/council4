package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dragon
//
// Type: Token Creature — Dragon
//
// Oracle text:
//   Flying

// DragonToken606c46d718a54cce9975029899b82b34 is the card definition for Dragon.
var DragonToken606c46d718a54cce9975029899b82b34 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dragon",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 6}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
