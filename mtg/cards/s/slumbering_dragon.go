package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlumberingDragon is the card definition for Slumbering Dragon.
//
// Type: Creature — Dragon
// Cost: {R}
//
// Oracle text:
//
//	Flying
//	This creature can't attack or block unless it has five or more +1/+1 counters on it.
//	Whenever a creature attacks you or a planeswalker you control, put a +1/+1 counter on this creature.
var SlumberingDragon = newSlumberingDragon

func newSlumberingDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Slumbering Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Negate:        true,
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}), RequiredCounter: counter.PlusOnePlusOne}),
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
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
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerYou,
							AttackRecipient:          game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker,
							SubjectSelection:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
							AttackRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerYou},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			This creature can't attack or block unless it has five or more +1/+1 counters on it.
			Whenever a creature attacks you or a planeswalker you control, put a +1/+1 counter on this creature.
		`,
		},
	}
}
