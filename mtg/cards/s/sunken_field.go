package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SunkenField is the card definition for Sunken Field.
//
// Type: Enchantment — Aura
// Cost: {1}{U}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}: Counter target spell unless its controller pays {1}."
var SunkenField = newSunkenField

func newSunkenField() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sunken Field",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:            "{T}: Counter target spell unless its controller pays {1}.",
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "target spell",
												Allow:      game.TargetAllowStackObject,
												Predicate: game.TargetPredicate{
													StackObjectKinds: []game.StackObjectKind{game.StackSpell},
												},
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.Pay{
													Payment: game.ResolutionPayment{
														Prompt: "Pay {1}?",
														Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
														ManaCost: opt.Val(cost.Mana{
															cost.O(1),
														}),
													},
												},
												PublishResult: game.ResultKey("unless-paid"),
											},
											{
												Primitive: game.CounterObject{
													Object: game.TargetStackObjectReference(0),
												},
												ResultGate: opt.Val(game.InstructionResultGate{
													Key:       "unless-paid",
													Succeeded: game.TriFalse,
												}),
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant land
			Enchanted land has "{T}: Counter target spell unless its controller pays {1}."
		`,
		},
	}
}
