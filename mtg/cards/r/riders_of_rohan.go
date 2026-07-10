package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RidersOfRohan is the card definition for Riders of Rohan.
//
// Type: Creature — Human Knight
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	When this creature enters, create two 2/2 red Human Knight creature tokens with trample and haste.
//	Dash {4}{R}{W} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var RidersOfRohan = newRidersOfRohan

func newRidersOfRohan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Riders of Rohan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(ridersOfRohanToken),
								},
							},
						},
					}.Ability(),
				},
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(4), cost.R, cost.W}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			When this creature enters, create two 2/2 red Human Knight creature tokens with trample and haste.
			Dash {4}{R}{W} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}

var ridersOfRohanToken = newRidersOfRohanToken()

func newRidersOfRohanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Knight",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
