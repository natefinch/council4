package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BoromirWardenOfTheTower is the card definition for Boromir, Warden of the Tower.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	Vigilance
//	Whenever an opponent casts a spell, if no mana was spent to cast it, counter that spell.
//	Sacrifice Boromir: Creatures you control gain indestructible until end of turn. The Ring tempts you.
var BoromirWardenOfTheTower = newBoromirWardenOfTheTower

func newBoromirWardenOfTheTower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Boromir, Warden of the Tower",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice Boromir: Creatures you control gain indestructible until end of turn. The Ring tempts you.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice Boromir",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.RingTempts{
									Player: game.ControllerReference(),
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
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerOpponent,
						},
						InterveningIf: "if no mana was spent to cast it",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventSpellManaSpentToCast, Op: compare.LessOrEqual, Value: 0}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.EventStackObjectReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			Whenever an opponent casts a spell, if no mana was spent to cast it, counter that spell.
			Sacrifice Boromir: Creatures you control gain indestructible until end of turn. The Ring tempts you.
		`,
		},
	}
}
