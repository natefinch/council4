package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuagmireLamprey is the card definition for Quagmire Lamprey.
//
// Type: Creature — Fish
// Cost: {2}{B}
//
// Oracle text:
//
//	Whenever this creature becomes blocked by a creature, put a -1/-1 counter on that creature.
var QuagmireLamprey = newQuagmireLamprey()

func newQuagmireLamprey() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Quagmire Lamprey",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerBecameBlocked,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventRelatedPermanentReference(),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature becomes blocked by a creature, put a -1/-1 counter on that creature.
		`,
		},
	}
}
