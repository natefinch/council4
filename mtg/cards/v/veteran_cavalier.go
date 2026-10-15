package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VeteranCavalier is the card definition for Veteran Cavalier.
//
// Type: Creature — Human Knight
// Cost: {W}{W}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
var VeteranCavalier = newVeteranCavalier

func newVeteranCavalier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Veteran Cavalier",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance (Attacking doesn't cause this creature to tap.)
		`,
		},
	}
}
