package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GreatSableStag is the card definition for Great Sable Stag.
//
// Type: Creature — Elk
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Protection from blue and from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything blue or black.)
var GreatSableStag = newGreatSableStag

func newGreatSableStag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Great Sable Stag",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elk},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Blue, color.Black),
			},
			OracleText: `
			This spell can't be countered.
			Protection from blue and from black (This creature can't be blocked, targeted, dealt damage, or enchanted by anything blue or black.)
		`,
		},
	}
}
