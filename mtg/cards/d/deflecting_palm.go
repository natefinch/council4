package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeflectingPalm is the card definition for Deflecting Palm.
//
// Type: Instant
// Cost: {R}{W}
//
// Oracle text:
//
//	The next time a source of your choice would deal damage to you this turn, prevent that damage. If damage is prevented this way, Deflecting Palm deals that much damage to that source's controller.
var DeflectingPalm = newDeflectingPalm

func newDeflectingPalm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Deflecting Palm",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							Player:                              game.ControllerReference(),
							All:                                 true,
							OneShot:                             true,
							RedirectPreventedToSourceController: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			The next time a source of your choice would deal damage to you this turn, prevent that damage. If damage is prevented this way, Deflecting Palm deals that much damage to that source's controller.
		`,
		},
	}
}
