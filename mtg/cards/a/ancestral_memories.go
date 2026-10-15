package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncestralMemories is the card definition for Ancestral Memories.
//
// Type: Sorcery
// Cost: {2}{U}{U}{U}
//
// Oracle text:
//
//	Look at the top seven cards of your library. Put two of them into your hand and the rest into your graveyard.
var AncestralMemories = newAncestralMemories

func newAncestralMemories() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ancestral Memories",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player: game.ControllerReference(),
							Look:   game.Fixed(7),
							Take:   game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top seven cards of your library. Put two of them into your hand and the rest into your graveyard.
		`,
		},
	}
}
