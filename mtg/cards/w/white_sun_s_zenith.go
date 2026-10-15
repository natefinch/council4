package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WhiteSunSZenith is the card definition for White Sun's Zenith.
//
// Type: Instant
// Cost: {X}{W}{W}{W}
//
// Oracle text:
//
//	Create X 2/2 white Cat creature tokens. Shuffle White Sun's Zenith into its owner's library.
var WhiteSunSZenith = newWhiteSunSZenith

func newWhiteSunSZenith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "White Sun's Zenith",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(whiteSunSZenithToken),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Library,
							Shuffle:     true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 2/2 white Cat creature tokens. Shuffle White Sun's Zenith into its owner's library.
		`,
		},
	}
}

var whiteSunSZenithToken = newWhiteSunSZenithToken()

func newWhiteSunSZenithToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
