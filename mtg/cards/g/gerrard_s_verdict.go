package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GerrardSVerdict is the card definition for Gerrard's Verdict.
//
// Type: Sorcery
// Cost: {W}{B}
//
// Oracle text:
//
//	Target player discards two cards. You gain 3 life for each land card discarded this way.
var GerrardSVerdict = newGerrardSVerdict

func newGerrardSVerdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Gerrard's Verdict",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
						PublishResult: game.ResultKey("actual-card-quantity-1"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 3,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Land}},
								ResultKey:  game.ResultKey("actual-card-quantity-1"),
							}),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "actual-card-quantity-1",
							Accepted: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards two cards. You gain 3 life for each land card discarded this way.
		`,
		},
	}
}
