package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FranklinRichardsAscendant is the card definition for Franklin Richards, Ascendant.
//
// Type: Legendary Creature — Mutant Hero
// Cost: {5}{R}
//
// Oracle text:
//
//	At the beginning of combat on your turn, if you've cast a noncreature spell this turn, discover 6. (Exile cards from the top of your library until you exile a nonland card with mana value 6 or less. Cast it without paying its mana cost or put it into your hand. Put the rest on the bottom in a random order.)
var FranklinRichardsAscendant = newFranklinRichardsAscendant

func newFranklinRichardsAscendant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Franklin Richards, Ascendant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Hero},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of combat on your turn, if you've cast a noncreature spell this turn, discover 6. (Exile cards from the top of your library until you exile a nonland card with mana value 6 or less. Cast it without paying its mana cost or put it into your hand. Put the rest on the bottom in a random order.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
						InterveningIf: "if you've cast a noncreature spell this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if you've cast a noncreature spell this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:           game.EventSpellCast,
									Controller:      game.TriggerControllerYou,
									CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
									StackObjectKind: game.StackSpell,
									CounterKind:     counter.PlusOnePlusOne,
								},
								Window: game.EventHistoryCurrentTurn,
							}),
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.DiscoverCards{
											Amount: game.Fixed(6),
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
			At the beginning of combat on your turn, if you've cast a noncreature spell this turn, discover 6. (Exile cards from the top of your library until you exile a nonland card with mana value 6 or less. Cast it without paying its mana cost or put it into your hand. Put the rest on the bottom in a random order.)
		`,
		},
	}
}
