package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FinchFormation is the card definition for FinchFormation.
//
// Type: Creature — Bird Scout
// Cost: {2}{U}
//
// Oracle text:
//
//	Offspring {3} (You may pay an additional {3} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Flying
//	When this creature enters, target creature you control gains flying until end of turn.
var FinchFormation = newFinchFormation

func newFinchFormation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Finch Formation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Scout},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(3)}),
				game.FlyingStaticBody,
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
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {3} (You may pay an additional {3} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Flying
			When this creature enters, target creature you control gains flying until end of turn.
		`,
		},
	}
}
