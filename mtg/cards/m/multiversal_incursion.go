package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MultiversalIncursion is the card definition for Multiversal Incursion.
//
// Type: Sorcery
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	For each nontoken creature you control, create a token that's a copy of that creature, except it isn't legendary.
var MultiversalIncursion = newMultiversalIncursion

func newMultiversalIncursion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Multiversal Incursion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:          game.TokenCopySourceEachInGroup,
								Group:           game.GroupRef(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, NonToken: true})),
								SetNotLegendary: true,
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each nontoken creature you control, create a token that's a copy of that creature, except it isn't legendary.
		`,
		},
	}
}
