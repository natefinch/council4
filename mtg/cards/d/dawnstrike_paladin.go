package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DawnstrikePaladin is the card definition for Dawnstrike Paladin.
//
// Type: Creature — Human Knight
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var DawnstrikePaladin = newDawnstrikePaladin

func newDawnstrikePaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dawnstrike Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Vigilance (Attacking doesn't cause this creature to tap.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
