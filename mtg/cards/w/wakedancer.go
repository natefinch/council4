package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wakedancer is the card definition for Wakedancer.
//
// Type: Creature — Human Shaman
// Cost: {2}{B}
//
// Oracle text:
//
//	Morbid — When this creature enters, if a creature died this turn, create a 2/2 black Zombie creature token.
var Wakedancer = newWakedancer

func newWakedancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Wakedancer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Morbid — When this creature enters, if a creature died this turn, create a 2/2 black Zombie creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if a creature died this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if a creature died this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									StackObjectKind:  game.StackSpell,
									CounterKind:      counter.PlusOnePlusOne,
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
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(wakedancerToken),
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
			Morbid — When this creature enters, if a creature died this turn, create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var wakedancerToken = newWakedancerToken()

func newWakedancerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
