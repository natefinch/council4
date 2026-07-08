package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhinoBarrelingBrute is the card definition for Rhino, Barreling Brute.
//
// Type: Legendary Creature — Human Villain
// Cost: {3}{R}{R}{G}{G}
//
// Oracle text:
//
//	Vigilance, trample, haste
//	Whenever Rhino attacks, if you've cast a spell with mana value 4 or greater this turn, draw a card.
var RhinoBarrelingBrute = newRhinoBarrelingBrute

func newRhinoBarrelingBrute() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Rhino, Barreling Brute",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Villain},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you've cast a spell with mana value 4 or greater this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:         game.EventSpellCast,
								Controller:    game.TriggerControllerYou,
								CardSelection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, trample, haste
			Whenever Rhino attacks, if you've cast a spell with mana value 4 or greater this turn, draw a card.
		`,
		},
	}
}
