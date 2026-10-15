package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommuneWithLava is the card definition for Commune with Lava.
//
// Type: Instant
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	Exile the top X cards of your library. Until the end of your next turn, you may play those cards.
var CommuneWithLava = newCommuneWithLava

func newCommuneWithLava() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Commune with Lava",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ImpulseExile{
							Player: game.ControllerReference(),
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile the top X cards of your library. Until the end of your next turn, you may play those cards.
		`,
		},
	}
}
