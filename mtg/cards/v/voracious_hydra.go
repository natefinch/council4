package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VoraciousHydra is the card definition for Voracious Hydra.
//
// Type: Creature — Hydra
// Cost: {X}{G}{G}
//
// Oracle text:
//
//	Trample
//	This creature enters with X +1/+1 counters on it.
//	When this creature enters, choose one —
//	• Double the number of +1/+1 counters on this creature.
//	• This creature fights target creature you don't control.
var VoraciousHydra = newVoraciousHydra

func newVoraciousHydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Voracious Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Double the number of +1/+1 counters on this creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												CounterKind: counter.PlusOnePlusOne,
												Object:      game.SourcePermanentReference(),
											}),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "This creature fights target creature you don't control.",
								Targets: []game.TargetSpec{
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
											Object:        game.SourcePermanentReference(),
											RelatedObject: game.TargetPermanentReference(0),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			Trample
			This creature enters with X +1/+1 counters on it.
			When this creature enters, choose one —
			• Double the number of +1/+1 counters on this creature.
			• This creature fights target creature you don't control.
		`,
		},
	}
}
