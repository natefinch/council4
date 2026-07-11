package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DevastatingSummons is the card definition for Devastating Summons.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice X lands.
//	Create two X/X red Elemental creature tokens.
var DevastatingSummons = newDevastatingSummons

func newDevastatingSummons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Devastating Summons",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice X lands",
					AmountFromX:        true,
					MatchPermanentType: true,
					PermanentType:      types.Land,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(devastatingSummonsToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice X lands.
			Create two X/X red Elemental creature tokens.
		`,
		},
	}
}

var devastatingSummonsToken = newDevastatingSummonsToken()

func newDevastatingSummonsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Elemental",
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Elemental},
		},
	}
}
