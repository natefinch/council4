package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlowOfKnowledge is the card definition for Flow of Knowledge.
//
// Type: Instant
// Cost: {4}{U}
//
// Oracle text:
//
//	Draw a card for each Island you control, then discard two cards.
var FlowOfKnowledge = newFlowOfKnowledge

func newFlowOfKnowledge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Flow of Knowledge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card for each Island you control, then discard two cards.
		`,
		},
	}
}
