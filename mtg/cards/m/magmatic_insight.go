package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MagmaticInsight is the card definition for Magmatic Insight.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a land card.
//	Draw two cards.
var MagmaticInsight = newMagmaticInsight

func newMagmaticInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Magmatic Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalDiscard,
					Text:          "discard a land card",
					Amount:        1,
					Source:        zone.Hand,
					MatchCardType: true,
					CardType:      types.Land,
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
			As an additional cost to cast this spell, discard a land card.
			Draw two cards.
		`,
		},
	}
}
