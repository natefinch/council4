package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GilraenDúnedainProtector is the card definition for Gilraen, Dúnedain Protector.
//
// Type: Legendary Creature — Human Noble
// Cost: {2}{W}
//
// Oracle text:
//
//	{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it.
var GilraenDúnedainProtector = newGilraenDúnedainProtector

func newGilraenDúnedainProtector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gilraen, Dúnedain Protector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.TargetPermanentReference(0),
									ExileLinkedKey: game.LinkedKey("blink-1"),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:       game.DelayedAtBeginningOfNextEndStep,
										CapturedCard: opt.Val(game.LinkedObjectReference("blink-1")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source:            game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
														EntryCounters:     []game.CounterPlacement{game.CounterPlacement{Kind: counter.Vigilance, Amount: 1}, game.CounterPlacement{Kind: counter.Lifelink, Amount: 1}},
														LinkedReturnZones: []zone.Type{zone.Exile},
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it.
		`,
		},
	}
}
