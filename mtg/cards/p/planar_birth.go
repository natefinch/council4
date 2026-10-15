package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlanarBirth is the card definition for Planar Birth.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	Return all basic land cards from all graveyards to the battlefield tapped under their owners' control.
var PlanarBirth = newPlanarBirth

func newPlanarBirth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Planar Birth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:            game.ControllerReference(),
							Selection:         game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
							Destination:       zone.Battlefield,
							EntryTapped:       true,
							SourceGroup:       game.AllPlayersReference(),
							ControlledByOwner: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all basic land cards from all graveyards to the battlefield tapped under their owners' control.
		`,
		},
	}
}
