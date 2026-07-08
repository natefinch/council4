package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FreestriderLookout is the card definition for Freestrider Lookout.
//
// Type: Creature — Human Rogue
// Cost: {2}{G}
//
// Oracle text:
//
//	Reach
//	Whenever you commit a crime, look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order. This ability triggers only once each turn. (Targeting opponents, anything they control, and/or cards in their graveyards is a crime.)
var FreestriderLookout = newFreestriderLookout

func newFreestriderLookout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Freestrider Lookout",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventCrimeCommitted,
							Player: game.TriggerPlayerYou,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:       game.ControllerReference(),
									Look:         game.Fixed(5),
									Take:         game.Fixed(1),
									Remainder:    game.DigRemainderLibraryBottom,
									Filter:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
									TakeUpTo:     true,
									Destination:  zone.Battlefield,
									EntersTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			Whenever you commit a crime, look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order. This ability triggers only once each turn. (Targeting opponents, anything they control, and/or cards in their graveyards is a crime.)
		`,
		},
	}
}
