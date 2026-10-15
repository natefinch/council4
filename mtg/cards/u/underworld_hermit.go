package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnderworldHermit is the card definition for Underworld Hermit.
//
// Type: Creature — Human Peasant
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	When this creature enters, create a number of 1/1 green Squirrel creature tokens equal to your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
var UnderworldHermit = newUnderworldHermit

func newUnderworldHermit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Underworld Hermit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Peasant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, create a number of 1/1 green Squirrel creature tokens equal to your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountDevotion,
												Multiplier: 1,
												Colors:     []color.Color{color.Black},
											}),
											Source: game.TokenDef(underworldHermitToken),
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
			When this creature enters, create a number of 1/1 green Squirrel creature tokens equal to your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
		`,
		},
	}
}

var underworldHermitToken = newUnderworldHermitToken()

func newUnderworldHermitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Squirrel",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
