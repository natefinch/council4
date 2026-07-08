package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShieldSphere is the card definition for Shield Sphere.
//
// Type: Artifact Creature — Wall
// Cost: {0}
//
// Oracle text:
//
//	Defender
//	Whenever this creature blocks, put a -0/-1 counter on it.
var ShieldSphere = newShieldSphere

func newShieldSphere() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Shield Sphere",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventBlockerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.MinusZeroMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			Whenever this creature blocks, put a -0/-1 counter on it.
		`,
		},
	}
}
