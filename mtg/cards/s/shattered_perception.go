package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShatteredPerception is the card definition for Shattered Perception.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Discard all the cards in your hand, then draw that many cards.
//	Flashback {5}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ShatteredPerception = newShatteredPerception

func newShatteredPerception() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Shattered Perception",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Text: "Discard all the cards in your hand, then draw that many cards.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Player:     game.ControllerReference(),
							EntireHand: true,
						},
						PublishResult: game.ResultKey("discarded-this-way"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("discarded-this-way"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Discard all the cards in your hand, then draw that many cards.
			Flashback {5}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
