package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyrantSChoice is the card definition for Tyrant's Choice.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Will of the council — Starting with you, each player votes for death or torture. If death gets more votes, each opponent sacrifices a creature of their choice. If torture gets more votes or the vote is tied, each opponent loses 4 life.
var TyrantSChoice = newTyrantSChoice

func newTyrantSChoice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tyrant's Choice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Vote{
							Options: []string{"death", "torture"},
						},
						PublishResult: game.ResultKey("vote-result"),
					},
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.OpponentsReference(),
							Amount:      game.Fixed(1),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "vote-result",
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 4}),
						}),
					},
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(4),
							PlayerGroup: game.OpponentsReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "vote-result",
							AmountRange: opt.Val(game.IntRange{Min: -4, Max: 0}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Will of the council — Starting with you, each player votes for death or torture. If death gets more votes, each opponent sacrifices a creature of their choice. If torture gets more votes or the vote is tied, each opponent loses 4 life.
		`,
		},
	}
}
