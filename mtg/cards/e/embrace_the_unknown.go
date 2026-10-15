package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmbraceTheUnknown is the card definition for Embrace the Unknown.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var EmbraceTheUnknown = newEmbraceTheUnknown

func newEmbraceTheUnknown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Embrace the Unknown",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ImpulseExile{
							Player:   game.ControllerReference(),
							Amount:   game.Fixed(2),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}
