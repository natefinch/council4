package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FaramirStewardOfGondor is the card definition for Faramir, Steward of Gondor.
//
// Type: Legendary Creature — Human Noble
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Whenever a legendary creature you control with mana value 4 or greater enters, you become the monarch.
//	At the beginning of your end step, if you're the monarch, create two 1/1 white Human Soldier creature tokens.
var FaramirStewardOfGondor = newFaramirStewardOfGondor

func newFaramirStewardOfGondor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Faramir, Steward of Gondor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if you're the monarch",
						InterveningCondition: opt.Val(game.Condition{
							ControllerIsMonarch: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(faramirStewardOfGondorToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a legendary creature you control with mana value 4 or greater enters, you become the monarch.
			At the beginning of your end step, if you're the monarch, create two 1/1 white Human Soldier creature tokens.
		`,
		},
	}
}

var faramirStewardOfGondorToken = newFaramirStewardOfGondorToken()

func newFaramirStewardOfGondorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
