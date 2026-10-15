package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FoundryAssembler is the card definition for Foundry Assembler.
//
// Type: Artifact Creature — Assembly-Worker
// Cost: {5}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
var FoundryAssembler = newFoundryAssembler

func newFoundryAssembler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Foundry Assembler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.AssemblyWorker},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
			},
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
		`,
		},
	}
}
