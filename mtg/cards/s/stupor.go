package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Stupor is the card definition for Stupor.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target opponent discards a card at random, then discards a card.
var Stupor = newStupor()

func newStupor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Stupor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent discards a card at random, then discards a card.
		`,
		},
	}
}
