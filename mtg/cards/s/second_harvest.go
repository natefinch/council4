package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SecondHarvest is the card definition for Second Harvest.
//
// Type: Instant
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	For each token you control, create a token that's a copy of that permanent.
var SecondHarvest = newSecondHarvest

func newSecondHarvest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Second Harvest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceEachInGroup,
								Group:  game.GroupRef(game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou, TokenOnly: true})),
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each token you control, create a token that's a copy of that permanent.
		`,
		},
	}
}
