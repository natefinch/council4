// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTCounterQuantitySpell is the card definition for RT Counter Quantity Spell.
//
// Type: Sorcery
//
// Oracle text:
//
//	Remove all +1/+1 counters from target creature. Draw that many cards.
var RTCounterQuantitySpell = newRTCounterQuantitySpell

func newRTCounterQuantitySpell() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Counter Quantity Spell",
			Types: []types.Card{types.Sorcery},
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
						Primitive: game.RemoveCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								CounterKind: counter.PlusOnePlusOne,
								Object:      game.TargetPermanentReference(0),
							}),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						PublishResult: game.ResultKey("removed-counter-quantity-1"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("removed-counter-quantity-1"),
							}),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "removed-counter-quantity-1",
							AmountAvailable: true,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Remove all +1/+1 counters from target creature. Draw that many cards.
		`,
		},
	}
}
