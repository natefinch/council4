package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MasterSCall is the card definition for Master's Call.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Create two 1/1 colorless Myr artifact creature tokens.
var MasterSCall = newMasterSCall

func newMasterSCall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Master's Call",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(masterSCallToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 colorless Myr artifact creature tokens.
		`,
		},
	}
}

var masterSCallToken = newMasterSCallToken()

func newMasterSCallToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Myr",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
