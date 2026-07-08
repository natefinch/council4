package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DinrovaHorror is the card definition for Dinrova Horror.
//
// Type: Creature — Horror
// Cost: {4}{U}{B}
//
// Oracle text:
//
//	When this creature enters, return target permanent to its owner's hand, then that player discards a card.
var DinrovaHorror = newDinrovaHorror

func newDinrovaHorror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Dinrova Horror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, return target permanent to its owner's hand, then that player discards a card.
		`,
		},
	}
}
