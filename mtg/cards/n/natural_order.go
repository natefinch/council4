package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NaturalOrder is the card definition for Natural Order.
//
// Type: Sorcery
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a green creature.
//	Search your library for a green creature card, put it onto the battlefield, then shuffle.
var NaturalOrder = newNaturalOrder

func newNaturalOrder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Natural Order",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a green creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
					MatchCardColor:     true,
					CardColor:          color.Green,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}},
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a green creature.
			Search your library for a green creature card, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
