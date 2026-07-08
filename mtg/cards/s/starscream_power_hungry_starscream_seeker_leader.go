package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StarscreamPowerHungry is the card definition for Starscream, Power Hungry // Starscream, Seeker Leader.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Starscream, Seeker Leader — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {2}{B} (You may cast this card converted for {2}{B}.)
//	Flying
//	Whenever you draw a card, if you're the monarch, target opponent loses 2 life.
//	Whenever one or more creatures deal combat damage to you, convert Starscream.
var StarscreamPowerHungry = newStarscreamPowerHungry()

func newStarscreamPowerHungry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Starscream, Power Hungry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventCardDrawn,
							Player: game.TriggerPlayerYou,
						},
						InterveningIf: "if you're the monarch",
						InterveningCondition: opt.Val(game.Condition{
							ControllerIsMonarch: true,
						}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Subject:               game.TriggerSubjectDamageSource,
							Player:                game.TriggerPlayerYou,
							OneOrMore:             true,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {2}{B} (You may cast this card converted for {2}{B}.)
			Flying
			Whenever you draw a card, if you're the monarch, target opponent loses 2 life.
			Whenever one or more creatures deal combat damage to you, convert Starscream.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Starscream, Seeker Leader",
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.FlyingStaticBody,
				game.MenaceStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
						InterveningIf: "if there is no monarch",
						InterveningCondition: opt.Val(game.Condition{
							NoMonarch: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventBecameMonarch,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Flying, menace, haste
			Whenever Starscream deals combat damage to a player, if there is no monarch, that player becomes the monarch.
			Whenever you become the monarch, convert Starscream.
		`,
		}),
	}
}
