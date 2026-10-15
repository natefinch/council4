package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RiseOfTheDarkRealms is the card definition for Rise of the Dark Realms.
//
// Type: Sorcery
// Cost: {7}{B}{B}
//
// Oracle text:
//
//	Put all creature cards from all graveyards onto the battlefield under your control.
var RiseOfTheDarkRealms = newRiseOfTheDarkRealms

func newRiseOfTheDarkRealms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rise of the Dark Realms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Destination: zone.Battlefield,
							SourceGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put all creature cards from all graveyards onto the battlefield under your control.
		`,
		},
	}
}
