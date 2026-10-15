package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GaeaSCourser is the card definition for Gaea's Courser.
//
// Type: Creature — Centaur Soldier
// Cost: {4}{G}
//
// Oracle text:
//
//	Whenever this creature attacks, if there are three or more creature cards in your graveyard, draw a card.
var GaeaSCourser = newGaeaSCourser

func newGaeaSCourser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gaea's Courser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Centaur, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, if there are three or more creature cards in your graveyard, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if there are three or more creature cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if there are three or more creature cards in your graveyard",
							SourceCounterKind: counter.PlusOnePlusOne,
							ControllerGraveyardCardOfTypeCountAtLeast: 3,
							ControllerGraveyardCountCardType:          types.Creature,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever this creature attacks, if there are three or more creature cards in your graveyard, draw a card.
		`,
		},
	}
}
