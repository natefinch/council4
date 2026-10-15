package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DarkPetition is the card definition for Dark Petition.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Search your library for a card, put that card into your hand, then shuffle.
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, add {B}{B}{B}.
var DarkPetition = newDarkPetition

func newDarkPetition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dark Petition",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
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
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate: "condition-0",
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a card, put that card into your hand, then shuffle.
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, add {B}{B}{B}.
		`,
		},
	}
}
