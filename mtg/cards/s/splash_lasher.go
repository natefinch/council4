package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SplashLasher is the card definition for SplashLasher.
//
// Type: Creature — Frog Wizard
// Cost: {3}{U}
//
// Oracle text:
//
//	Offspring {1}{U} (You may pay an additional {1}{U} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	When this creature enters, tap up to one target creature and put a stun counter on it. (If a permanent with a stun counter would become untapped, remove one from it instead.)
var SplashLasher = newSplashLasher

func newSplashLasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Splash Lasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(1), cost.U}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Stun,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {1}{U} (You may pay an additional {1}{U} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			When this creature enters, tap up to one target creature and put a stun counter on it. (If a permanent with a stun counter would become untapped, remove one from it instead.)
		`,
		},
	}
}
