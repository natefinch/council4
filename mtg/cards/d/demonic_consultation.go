package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DemonicConsultation is the card definition for Demonic Consultation.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Choose a card name. Exile the top six cards of your library, then reveal cards from the top of your library until you reveal a card with the chosen name. Put that card into your hand and exile all other cards revealed this way.
var DemonicConsultation = newDemonicConsultation

func newDemonicConsultation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Demonic Consultation",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.IterativeLibraryProcess{
							Player:          game.ControllerReference(),
							Stop:            game.IterativeLibraryStopChosenName,
							PreExile:        game.Fixed(6),
							ChooseName:      true,
							Reveal:          true,
							AllowAbsentName: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a card name. Exile the top six cards of your library, then reveal cards from the top of your library until you reveal a card with the chosen name. Put that card into your hand and exile all other cards revealed this way.
		`,
		},
	}
}
