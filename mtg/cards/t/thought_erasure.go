package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThoughtErasure is the card definition for Thought Erasure.
//
// Type: Sorcery
// Cost: {U}{B}
//
// Oracle text:
//
//	Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.
//	Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
var ThoughtErasure = newThoughtErasure

func newThoughtErasure() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Thought Erasure",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseDiscardFromHand{
							Player:      game.TargetPlayerReference(0),
							ExcludeLand: true,
						},
					},
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.
			Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
		`,
		},
	}
}
