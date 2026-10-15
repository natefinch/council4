package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DiabolicIntent is the card definition for Diabolic Intent.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Search your library for a card, put that card into your hand, then shuffle.
var DiabolicIntent = newDiabolicIntent

func newDiabolicIntent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Diabolic Intent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
								SourceZone:       zone.Library,
								Destination:      zone.Hand,
								FailToFindPolicy: game.SearchMustFindIfAvailable,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Search your library for a card, put that card into your hand, then shuffle.
		`,
		},
	}
}
