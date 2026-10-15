package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArtificerSEpiphany is the card definition for Artificer's Epiphany.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Draw two cards. If you control no artifacts, discard a card.
var ArtificerSEpiphany = newArtificerSEpiphany

func newArtificerSEpiphany() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Artificer's Epiphany",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
									MinCount:  1,
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards. If you control no artifacts, discard a card.
		`,
		},
	}
}
