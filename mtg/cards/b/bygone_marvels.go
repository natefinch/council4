package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BygoneMarvels is the card definition for Bygone Marvels.
//
// Type: Sorcery
// Cost: {G}{G}
//
// Oracle text:
//
//	Descend 8 — When you cast this spell, if there are eight or more permanent cards in your graveyard, copy this spell twice. You may choose new targets for the copies.
//	Return target permanent card from your graveyard to your hand. Exile Bygone Marvels.
var BygoneMarvels = newBygoneMarvels

func newBygoneMarvels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bygone Marvels",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Descend 8 — When you cast this spell, if there are eight or more permanent cards in your graveyard, copy this spell twice. You may choose new targets for the copies.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							Source:          game.TriggerSourceSelf,
							SelfWasCast:     true,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if there are eight or more permanent cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if there are eight or more permanent cards in your graveyard",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateControllerGraveyardPermanentCardCount,
									Op:        compare.GreaterOrEqual,
									Value:     8,
								},
							},
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object:              game.ResolvingStackObjectReference(),
											Count:               2,
											MayChooseNewTargets: true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Descend 8 — When you cast this spell, if there are eight or more permanent cards in your graveyard, copy this spell twice. You may choose new targets for the copies.
			Return target permanent card from your graveyard to your hand. Exile Bygone Marvels.
		`,
		},
	}
}
