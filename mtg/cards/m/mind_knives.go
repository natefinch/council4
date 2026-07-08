package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindKnives is the card definition for Mind Knives.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target opponent discards a card at random.
var MindKnives = newMindKnives

func newMindKnives() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mind Knives",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
				},
			}.Ability()),
			OracleText: `
			Target opponent discards a card at random.
		`,
		},
	}
}
