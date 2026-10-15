package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VenomousChangeling is the card definition for Venomous Changeling.
//
// Type: Creature — Shapeshifter
// Cost: {2}{B}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Deathtouch
var VenomousChangeling = newVenomousChangeling

func newVenomousChangeling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Venomous Changeling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Deathtouch
		`,
		},
	}
}
