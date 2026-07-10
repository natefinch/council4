package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InvasionOfKylem is the card definition for Invasion of Kylem // Valor's Reach Tag Team.
//
// Type: Battle — Siege // Sorcery
// Face: Valor's Reach Tag Team — Sorcery
//
// Oracle text:
//
//	(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
//	When this Siege enters, up to two target creatures each get +2/+0 and gain vigilance and haste until end of turn.
var InvasionOfKylem = newInvasionOfKylem

func newInvasionOfKylem() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Invasion of Kylem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.W,
			}),
			Colors:   []color.Color{color.Red, color.White},
			Types:    []types.Card{types.Battle},
			Subtypes: []types.Sub{types.Siege},
			Defense:  opt.Val(5),
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
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 2,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Vigilance,
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(1)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 2,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Vigilance,
												game.Haste,
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
			(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
			When this Siege enters, up to two target creatures each get +2/+0 and gain vigilance and haste until end of turn.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:   "Valor's Reach Tag Team",
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(invasionOfKylemToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 3/2 red and white Warrior creature tokens with "Whenever this token and at least one other creature token attack, put a +1/+1 counter on this token."
		`,
		}),
	}
}

var invasionOfKylemToken = newInvasionOfKylemToken()

func newInvasionOfKylemToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Warrior",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                     game.EventAttackerDeclared,
							Source:                    game.TriggerSourceSelf,
							AttacksAlongsideCount:     1,
							AttacksAlongsideSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
