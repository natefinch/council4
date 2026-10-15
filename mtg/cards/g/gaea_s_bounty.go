package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GaeaSBounty is the card definition for Gaea's Bounty.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Search your library for up to two Forest cards, reveal those cards, put them into your hand, then shuffle.
var GaeaSBounty = newGaeaSBounty

func newGaeaSBounty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gaea's Bounty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Forest}},
								Reveal:      true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to two Forest cards, reveal those cards, put them into your hand, then shuffle.
		`,
		},
	}
}
