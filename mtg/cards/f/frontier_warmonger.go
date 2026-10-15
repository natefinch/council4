package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrontierWarmonger is the card definition for Frontier Warmonger.
//
// Type: Creature — Human Warrior
// Cost: {3}{R}
//
// Oracle text:
//
//	Whenever one or more creatures attack one of your opponents or a planeswalker they control, those creatures gain menace until end of turn.
var FrontierWarmonger = newFrontierWarmonger

func newFrontierWarmonger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Frontier Warmonger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more creatures attack one of your opponents or a planeswalker they control, those creatures gain menace until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerOpponent,
							SubjectSelection:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:          game.StackSpell,
							AttackRecipient:          game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker,
							AttackRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerNotYou},
							OneOrMore:                true,
							CounterKind:              counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Group:       game.TriggeringAttackersAgainstDefenderGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.TriggerControllerOpponent),
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Menace},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Whenever one or more creatures attack one of your opponents or a planeswalker they control, those creatures gain menace until end of turn.
		`,
		},
	}
}
