package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RoonOfTheHiddenRealm is the card definition for Roon of the Hidden Realm.
//
// Type: Legendary Creature — Rhino Soldier
// Cost: {2}{G}{W}{U}
//
// Oracle text:
//
//	Vigilance, trample
//	{2}, {T}: Exile another target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.
var RoonOfTheHiddenRealm = newRoonOfTheHiddenRealm

func newRoonOfTheHiddenRealm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Roon of the Hidden Realm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Rhino, types.Soldier},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}: Exile another target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:        game.TargetPermanentReference(0),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										ClearLinkedBeforeGate: true,
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtBeginningOfNextEndStep,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.PutOnBattlefield{
																		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedCard: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
											},
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
			OracleText: `
			Vigilance, trample
			{2}, {T}: Exile another target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.
		`,
		},
	}
}
