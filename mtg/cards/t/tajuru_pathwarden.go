package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TajuruPathwarden is the card definition for Tajuru Pathwarden.
//
// Type: Creature — Elf Warrior Ally
// Cost: {4}{G}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
var TajuruPathwarden = newTajuruPathwarden

func newTajuruPathwarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tajuru Pathwarden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior, types.Ally},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Vigilance (Attacking doesn't cause this creature to tap.)
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
