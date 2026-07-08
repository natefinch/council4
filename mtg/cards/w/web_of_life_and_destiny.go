package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WebOfLifeAndDestiny is the card definition for Web of Life and Destiny.
//
// Type: Enchantment
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	At the beginning of combat on your turn, look at the top five cards of your library. You may put a creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
var WebOfLifeAndDestiny = newWebOfLifeAndDestiny

func newWebOfLifeAndDestiny() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Web of Life and Destiny",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:      game.ControllerReference(),
									Look:        game.Fixed(5),
									Take:        game.Fixed(1),
									Remainder:   game.DigRemainderLibraryBottom,
									Filter:      opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									TakeUpTo:    true,
									Destination: zone.Battlefield,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			At the beginning of combat on your turn, look at the top five cards of your library. You may put a creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
