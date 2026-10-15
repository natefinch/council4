package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JillShivaSDominant is the card definition for Jill, Shiva's Dominant // Shiva, Warden of Ice.
//
// Type: Legendary Creature — Human Noble Warrior // Legendary Enchantment Creature — Saga Elemental
// Face: Shiva, Warden of Ice — Legendary Enchantment Creature — Saga Elemental
//
// Oracle text:
//
//	When Jill enters, return up to one other target nonland permanent to its owner's hand.
//	{3}{U}{U}, {T}: Exile Jill, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.
var JillShivaSDominant = newJillShivaSDominant

func newJillShivaSDominant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jill, Shiva's Dominant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble, types.Warrior},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{U}{U}, {T}: Exile Jill, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.U,
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:        game.SourcePermanentReference(),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("self-blink-1"),
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:           game.LinkedBattlefieldSource(game.LinkedKey("self-blink-1")),
											EntryTransformed: true,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Jill enters, return up to one other target nonland permanent to its owner's hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one other target nonland permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludeSource: true}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Hand,
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
			When Jill enters, return up to one other target nonland permanent to its owner's hand.
			{3}{U}{U}, {T}: Exile Jill, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Shiva, Warden of Ice",
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.Saga, types.Elemental},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — Mesmerize — Target creature can't be blocked this turn.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.ApplyRule{
											Object: opt.Val(game.TargetPermanentReference(0)),
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:     game.RuleEffectCantBeBlocked,
													Duration: game.DurationPermanent,
												},
											},
											Duration: game.DurationThisTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Cold Snap — Tap all lands your opponents control. Exile Shiva, then return it to the battlefield (front face up).",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Tap{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent}),
										},
									},
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:        game.SourcePermanentReference(),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("self-blink-2"),
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.LinkedBattlefieldSource(game.LinkedKey("self-blink-2")),
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
			(As this Saga enters and after your draw step, add a lore counter.)
			I, II — Mesmerize — Target creature can't be blocked this turn.
			III — Cold Snap — Tap all lands your opponents control. Exile Shiva, then return it to the battlefield (front face up).
		`,
		}),
	}
}
