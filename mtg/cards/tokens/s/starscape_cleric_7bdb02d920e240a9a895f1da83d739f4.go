package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StarscapeCleric is the card definition for Starscape Cleric.
//
// Type: Token Creature — Bat Cleric
// Cost: {1}{B}
//
// Oracle text:
//   Flying
//   This creature can't block.
//   Whenever you gain life, each opponent loses 1 life.
//   (This token's mana cost is {1}{B}.)

// StarscapeClericToken7bdb02d920e240a9a895f1da83d739f4 is the card definition for Starscape Cleric.
var StarscapeClericToken7bdb02d920e240a9a895f1da83d739f4 = newStarscapeClericToken7bdb02d920e240a9a895f1da83d739f4()

func newStarscapeClericToken7bdb02d920e240a9a895f1da83d739f4() *game.CardDef {
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
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you gain life, each opponent loses 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventLifeGained,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
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
			Flying
			This creature can't block.
			Whenever you gain life, each opponent loses 1 life.
			(This token's mana cost is {1}{B}.)
		`,
		},
	}
}
