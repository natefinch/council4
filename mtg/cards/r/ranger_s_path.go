package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RangerSPath is the card definition for Ranger's Path.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Search your library for up to two Forest cards, put them onto the battlefield tapped, then shuffle.
var RangerSPath = newRangerSPath

func newRangerSPath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ranger's Path",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:   zone.Library,
								Destination:  zone.Battlefield,
								Filter:       game.Selection{SubtypesAny: []types.Sub{types.Forest}},
								EntersTapped: true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to two Forest cards, put them onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
