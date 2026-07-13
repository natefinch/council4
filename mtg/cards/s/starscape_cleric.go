package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StarscapeCleric is the card definition for StarscapeCleric.
//
// Type: Creature — Bat Cleric
// Cost: {1}{B}
//
// Oracle text:
//
//	Offspring {2}{B} (You may pay an additional {2}{B} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Flying
//	This creature can't block.
//	Whenever you gain life, each opponent loses 1 life.
var StarscapeCleric = newStarscapeCleric

func newStarscapeCleric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Starscape Cleric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bat, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(2), cost.B}),
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLifeGained,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {2}{B} (You may pay an additional {2}{B} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Flying
			This creature can't block.
			Whenever you gain life, each opponent loses 1 life.
		`,
		},
	}
}
