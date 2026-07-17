package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExpressiveIteration is the card definition for Expressive Iteration.
//
// Type: Sorcery
// Cost: {U}{R}
//
// Oracle text:
//
//	Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.
var ExpressiveIteration = newExpressiveIteration

func newExpressiveIteration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Expressive Iteration",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(3),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Slots: []game.DigSlot{
								game.DigSlot{
									Count:       game.Fixed(1),
									Destination: zone.Library,
									Bottom:      true,
								},
								game.DigSlot{
									Count:       game.Fixed(1),
									Destination: zone.Exile,
									Play: opt.Val(game.ImpulsePlayGrant{
										Duration: game.DurationThisTurn,
									}),
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.
		`,
		},
	}
}
