package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LifeGoesOn is the card definition for Life Goes On.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	You gain 4 life. If a creature died this turn, you gain 8 life instead.
var LifeGoesOn = newLifeGoesOn()

func newLifeGoesOn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Life Goes On",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(8),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			You gain 4 life. If a creature died this turn, you gain 8 life instead.
		`,
		},
	}
}
