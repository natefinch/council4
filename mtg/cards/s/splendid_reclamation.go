package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SplendidReclamation is the card definition for Splendid Reclamation.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Return all land cards from your graveyard to the battlefield tapped.
var SplendidReclamation = newSplendidReclamation

func newSplendidReclamation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Splendid Reclamation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all land cards from your graveyard to the battlefield tapped.
		`,
		},
	}
}
