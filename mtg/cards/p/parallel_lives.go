package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ParallelLives is the card definition for Parallel Lives.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
var ParallelLives = newParallelLives

func newParallelLives() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Parallel Lives",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
		`,
		},
	}
}
