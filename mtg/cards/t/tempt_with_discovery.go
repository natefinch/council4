package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TemptWithDiscovery is the card definition for Tempt with Discovery.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Tempting offer — Search your library for a land card and put it onto the battlefield. Each opponent may search their library for a land card and put it onto the battlefield. For each opponent who searches a library this way, search your library for a land card and put it onto the battlefield. Then each player who searched a library this way shuffles.
var TemptWithDiscovery = newTemptWithDiscovery

func newTemptWithDiscovery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tempt with Discovery",
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
							Player: game.GroupOfferMemberReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
							},
							Amount: game.Fixed(1),
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.OpponentsReference()),
						TemptingOffer:      true,
					},
				},
			}.Ability()),
			OracleText: `
			Tempting offer — Search your library for a land card and put it onto the battlefield. Each opponent may search their library for a land card and put it onto the battlefield. For each opponent who searches a library this way, search your library for a land card and put it onto the battlefield. Then each player who searched a library this way shuffles.
		`,
		},
	}
}
