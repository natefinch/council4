package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// QuarrelSEnd is the card definition for Quarrel's End.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card.
//	Draw two cards and create a 1/1 white Human Soldier creature token.
var QuarrelSEnd = newQuarrelSEnd

func newQuarrelSEnd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Quarrel's End",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
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
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(quarrelSEndToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card.
			Draw two cards and create a 1/1 white Human Soldier creature token.
		`,
		},
	}
}

var quarrelSEndToken = newQuarrelSEndToken()

func newQuarrelSEndToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
