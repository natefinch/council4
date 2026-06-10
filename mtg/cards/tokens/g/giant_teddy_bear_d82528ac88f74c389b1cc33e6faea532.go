package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Giant Teddy Bear
//
// Type: Token Creature — Giant Teddy Bear
//
// Oracle text:

// GiantTeddyBearTokend82528ac88f74c389b1cc33e6faea532 is the card definition for Giant Teddy Bear.
var GiantTeddyBearTokend82528ac88f74c389b1cc33e6faea532 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Giant Teddy Bear",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant, types.Sub("Teddy"), types.Bear},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	},
}
