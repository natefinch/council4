package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Beast
//
// Type: Token Artifact Creature — Beast
//
// Oracle text:

// BeastToken347a9b418545405d9b4b3691890201d7 is the card definition for Beast.
var BeastToken347a9b418545405d9b4b3691890201d7 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Beast",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 6}),
	},
}
