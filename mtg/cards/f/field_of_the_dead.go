package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieldOfTheDead is the card definition for Field of the Dead.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {C}.
//	Whenever this land or another land you control enters, if you control seven or more lands with different names, create a 2/2 black Zombie creature token.
var FieldOfTheDead = newFieldOfTheDead

func newFieldOfTheDead() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Field of the Dead",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this land or another land you control enters, if you control seven or more lands with different names, create a 2/2 black Zombie creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Land}},
							SubjectSelectionOrSelf: true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
						InterveningIf: "if you control seven or more lands with different names",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you control seven or more lands with different names",
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
								MinCount:  7,
								DistinctNames: opt.Val(compare.Int{
									Op:    compare.GreaterOrEqual,
									Value: 7,
								}),
							}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(fieldOfTheDeadToken),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {C}.
			Whenever this land or another land you control enters, if you control seven or more lands with different names, create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var fieldOfTheDeadToken = newFieldOfTheDeadToken()

func newFieldOfTheDeadToken() *game.CardDef {
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
