package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindTwist is the card definition for Mind Twist.
//
// Type: Sorcery
// Cost: {X}{B}
//
// Oracle text:
//
//	Target player discards X cards at random.
var MindTwist = newMindTwist

func newMindTwist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mind Twist",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards X cards at random.
		`,
		},
	}
}
