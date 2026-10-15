package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrackDown is the card definition for Track Down.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Scry 3, then reveal the top card of your library. If it's a creature or land card, draw a card. (To scry 3, look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var TrackDown = newTrackDown

func newTrackDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Track Down",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Reveal{
							Amount:        game.Fixed(1),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("sequence-effect-1-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-1-product")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-1-product")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Scry 3, then reveal the top card of your library. If it's a creature or land card, draw a card. (To scry 3, look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}
