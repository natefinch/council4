package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LaughingMad is the card definition for Laughing Mad.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card.
//	Draw two cards.
//	Flashback {3}{R} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
var LaughingMad = newLaughingMad

func newLaughingMad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Laughing Mad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.R,
							},
						},
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalDiscard,
					Text:   "discard a card",
					Amount: 1,
					Source: zone.Hand,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card.
			Draw two cards.
			Flashback {3}{R} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
		`,
		},
	}
}
