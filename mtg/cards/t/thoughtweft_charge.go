package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThoughtweftCharge is the card definition for Thoughtweft Charge.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature gets +3/+3 until end of turn. If a creature entered the battlefield under your control this turn, draw a card.
var ThoughtweftCharge = newThoughtweftCharge()

func newThoughtweftCharge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Thoughtweft Charge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(3),
							ToughnessDelta: game.Fixed(3),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventPermanentEnteredBattlefield,
									Controller:       game.TriggerControllerYou,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +3/+3 until end of turn. If a creature entered the battlefield under your control this turn, draw a card.
		`,
		},
	}
}
