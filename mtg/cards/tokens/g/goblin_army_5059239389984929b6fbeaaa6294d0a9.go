package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinArmy is the card definition for Goblin Army.
//
// Type: Token Creature — Goblin Army
//
// Oracle text:

// GoblinArmyToken5059239389984929b6fbeaaa6294d0a9 is the card definition for Goblin Army.
var GoblinArmyToken5059239389984929b6fbeaaa6294d0a9 = newGoblinArmyToken5059239389984929b6fbeaaa6294d0a9()

func newGoblinArmyToken5059239389984929b6fbeaaa6294d0a9() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin Army",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Army},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
