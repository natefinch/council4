package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IsaoEnlightenedBushi is the card definition for Isao, Enlightened Bushi.
//
// Type: Legendary Creature — Human Samurai
// Cost: {2}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)
//	{2}: Regenerate target Samurai.
var IsaoEnlightenedBushi = newIsaoEnlightenedBushi

func newIsaoEnlightenedBushi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Isao, Enlightened Bushi",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Samurai},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Regenerate target Samurai.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Samurai",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Samurai}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
											Object: game.TargetPermanentReference(0),
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
					Text: "Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventBlockerDeclared,
							BlockerScope:    game.BlockerTriggerOnce,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventAttackerBecameBlocked,
							BlockedScope:    game.BlockedTriggerOnce,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.EventPermanentReference(),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
											Duration:       game.DurationUntilEndOfTurn,
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
			This spell can't be countered.
			Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)
			{2}: Regenerate target Samurai.
		`,
		},
	}
}
