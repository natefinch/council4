package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InscriptionOfAbundance is the card definition for Inscription of Abundance.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Kicker {2}{G}
//	Choose one. If this spell was kicked, choose any number instead.
//	• Put two +1/+1 counters on target creature.
//	• Target player gains X life, where X is the greatest power among creatures they control.
//	• Target creature you control fights target creature you don't control.
var InscriptionOfAbundance = newInscriptionOfAbundance

func newInscriptionOfAbundance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Inscription of Abundance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.G}},
					},
				},
			},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Put two +1/+1 counters on target creature.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					},
					game.Mode{
						Text: "Target player gains X life, where X is the greatest power among creatures they control.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestPowerInGroup,
										Multiplier: 1,
										Group:      game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature you control fights target creature you don't control.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you don't control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Fight{
									Object:        game.TargetPermanentReference(0),
									RelatedObject: game.TargetPermanentReference(1),
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionSpellKicked, ReplaceRange: true, MinModes: 0, MaxModes: 3},
			}),
			OracleText: `
			Kicker {2}{G}
			Choose one. If this spell was kicked, choose any number instead.
			• Put two +1/+1 counters on target creature.
			• Target player gains X life, where X is the greatest power among creatures they control.
			• Target creature you control fights target creature you don't control.
		`,
		},
	}
}
