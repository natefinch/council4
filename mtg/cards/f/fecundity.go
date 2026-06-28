package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fecundity is the card definition for Fecundity.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Whenever a creature dies, that creature's controller may draw a card.
var Fecundity = newFecundity()

func newFecundity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fecundity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature dies, that creature's controller may draw a card.
		`,
		},
	}
}
