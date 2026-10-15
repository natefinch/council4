package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PuppetSVerdict is the card definition for Puppet's Verdict.
//
// Type: Instant
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Flip a coin. If you win the flip, destroy all creatures with power 2 or less. If you lose the flip, destroy all creatures with power 3 or greater.
var PuppetSVerdict = newPuppetSVerdict

func newPuppetSVerdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Puppet's Verdict",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RollDie{
							Sides: 2,
						},
						PublishResult: game.ResultKey("coin-flip-result"),
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "coin-flip-result",
							AmountRange: opt.Val(game.IntRange{Min: 2, Max: 2}),
						}),
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "coin-flip-result",
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 1}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Flip a coin. If you win the flip, destroy all creatures with power 2 or less. If you lose the flip, destroy all creatures with power 3 or greater.
		`,
		},
	}
}
