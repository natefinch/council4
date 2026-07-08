package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeaHag is the card definition for Sea Hag // Aquatic Ingress.
//
// Type: Creature — Hag // Instant — Adventure
// Cost: {4}{U} // {2}{U}
// Face: Aquatic Ingress — Instant — Adventure ({2}{U})
//
// Oracle text:
//
//	When this creature enters, creatures your opponents control get -4/-0 until end of turn.
var SeaHag = newSeaHag

func newSeaHag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sea Hag",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hag},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
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
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
											PowerDelta: -4,
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
			When this creature enters, creatures your opponents control get -4/-0 until end of turn.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Aquatic Ingress",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(1),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(1),
							PowerDelta:     game.Fixed(1),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind: game.RuleEffectCantBeBlocked,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(1)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind: game.RuleEffectCantBeBlocked,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Up to two target creatures each get +1/+0 until end of turn and can't be blocked this turn. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
