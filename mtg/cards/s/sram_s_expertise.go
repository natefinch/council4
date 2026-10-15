package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SramSExpertise is the card definition for Sram's Expertise.
//
// Type: Sorcery
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Create three 1/1 colorless Servo artifact creature tokens.
//	You may cast a spell with mana value 3 or less from your hand without paying its mana cost.
var SramSExpertise = newSramSExpertise

func newSramSExpertise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sram's Expertise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(sramSExpertiseToken),
						},
					},
					{
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 3,
							})},
							Zone: zone.Hand,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 colorless Servo artifact creature tokens.
			You may cast a spell with mana value 3 or less from your hand without paying its mana cost.
		`,
		},
	}
}

var sramSExpertiseToken = newSramSExpertiseToken()

func newSramSExpertiseToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Servo",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Servo},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
