package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EldritchEvolution is the card definition for Eldritch Evolution.
//
// Type: Sorcery
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Search your library for a creature card with mana value X or less, where X is 2 plus the sacrificed creature's mana value. Put that card onto the battlefield, then shuffle. Exile Eldritch Evolution.
var EldritchEvolution = newEldritchEvolution

func newEldritchEvolution() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Eldritch Evolution",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:                     zone.Library,
								Destination:                    zone.Battlefield,
								Filter:                         game.Selection{RequiredTypes: []types.Card{types.Creature}},
								MaxManaValueFromSacrificedCost: opt.Val(2),
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.Exile{
							SourceSpell: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Search your library for a creature card with mana value X or less, where X is 2 plus the sacrificed creature's mana value. Put that card onto the battlefield, then shuffle. Exile Eldritch Evolution.
		`,
		},
	}
}
