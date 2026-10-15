package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragTheCanal is the card definition for Drag the Canal.
//
// Type: Instant
// Cost: {U}{B}
//
// Oracle text:
//
//	Create a 2/2 white and blue Detective creature token. If a creature died this turn, you gain 2 life, surveil 2, then investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
var DragTheCanal = newDragTheCanal

func newDragTheCanal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Drag the Canal",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(dragTheCanalToken),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									StackObjectKind:  game.StackSpell,
									CounterKind:      counter.PlusOnePlusOne,
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.Investigate{
							Amount: game.Fixed(1),
						},
						ConditionGate: "condition-0",
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 white and blue Detective creature token. If a creature died this turn, you gain 2 life, surveil 2, then investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
		`,
		},
	}
}

var dragTheCanalToken = newDragTheCanalToken()

func newDragTheCanalToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Detective",
			Colors:    []color.Color{color.White, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Detective},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
