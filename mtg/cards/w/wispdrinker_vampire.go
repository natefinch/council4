package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WispdrinkerVampire is the card definition for Wispdrinker Vampire.
//
// Type: Creature — Vampire Rogue
// Cost: {2}{W}{B}
//
// Oracle text:
//
//	Flying
//	Whenever another creature you control with power 2 or less enters, each opponent loses 1 life and you gain 1 life.
//	{5}{W}{B}: Creatures you control with power 2 or less gain deathtouch and lifelink until end of turn.
var WispdrinkerVampire = newWispdrinkerVampire

func newWispdrinkerVampire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Wispdrinker Vampire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{5}{W}{B}: Creatures you control with power 2 or less gain deathtouch and lifelink until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(5), cost.W, cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
											AddKeywords: []game.Keyword{
												game.Deathtouch,
												game.Lifelink,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
								PublishResult: game.ResultKey("life-change"),
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever another creature you control with power 2 or less enters, each opponent loses 1 life and you gain 1 life.
			{5}{W}{B}: Creatures you control with power 2 or less gain deathtouch and lifelink until end of turn.
		`,
		},
	}
}
