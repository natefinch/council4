package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianProject is the card definition for Guardian Project.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.
var GuardianProject = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Guardian Project",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			OracleText: `
			Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.
		`,
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: `
					Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.
				`,
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Controller: game.TriggerControllerYou,
							RequirePermanentTypes: []types.Card{
								types.Creature,
							},
							RequireNonToken: true,
						},
						InterveningIf: "it doesn't have the same name as another creature you control or a creature card in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text: "it doesn't have the same name as another creature you control or a creature card in your graveyard",
							EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
