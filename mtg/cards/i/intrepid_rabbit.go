package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IntrepidRabbit is the card definition for IntrepidRabbit.
//
// Type: Creature — Rabbit Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	Offspring {1} (You may pay an additional {1} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	When this creature enters, target creature you control gets +1/+1 until end of turn.
var IntrepidRabbit = newIntrepidRabbit

func newIntrepidRabbit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Intrepid Rabbit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(1)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
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
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {1} (You may pay an additional {1} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			When this creature enters, target creature you control gets +1/+1 until end of turn.
		`,
		},
	}
}
