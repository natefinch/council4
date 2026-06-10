package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Assembly-Worker
//
// Type: Token Artifact Creature — Assembly-Worker
//
// Oracle text:

// AssemblyWorkerToken02e1a47136b24d28a323f4ec3d095cb9 is the card definition for Assembly-Worker.
var AssemblyWorkerToken02e1a47136b24d28a323f4ec3d095cb9 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Assembly-Worker",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.AssemblyWorker},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
