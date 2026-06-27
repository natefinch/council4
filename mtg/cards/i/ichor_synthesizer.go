package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IchorSynthesizer is the card definition for Ichor Synthesizer.
//
// Type: Creature — Phyrexian Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Whenever you cast a noncreature spell, put an oil counter on this creature.
//	As long as this creature has four or more oil counters on it, it gets +2/+0 and can't be blocked.
var IchorSynthesizer = newIchorSynthesizer()

func newIchorSynthesizer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ichor Synthesizer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}), RequiredCounter: counter.Oil}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Oil,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast a noncreature spell, put an oil counter on this creature.
			As long as this creature has four or more oil counters on it, it gets +2/+0 and can't be blocked.
		`,
		},
	}
}
